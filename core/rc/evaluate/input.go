// Package evaluate explains Remote Config expressions against captured client
// facts. It never sends, executes, or persists a captured request.
package evaluate

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const MaxCaptureBytes = 32 << 20

// Fact distinguishes a known value, an explicitly absent value, and information
// unavailable to the evaluator. Source is retained in the explanation.
type Fact struct {
	Value            string
	Known            bool
	Absent           bool
	Source           string
	Reason           string
	TimeMin, TimeMax time.Time
	DerivedFrom      []string
}

type Exchange struct {
	Format        string
	Entry         int
	Project       string
	Namespace     string
	StartedAt     time.Time
	Status        int
	Facts         map[string]Fact
	Entries       map[string]string
	Version       string
	HasResponse   bool
	Complete      map[string]bool
	TimeZone      string
	ResponseState string
	Managed       map[string]string
	Experiments   []string
	Assignments   map[string][]Assignment
}

type Assignment struct {
	Kind    string  `json:"kind" contract:"enum=personalization|rollout"`
	ID      *string `json:"id"`
	Variant *string `json:"variant"`
}

// CaptureError marks invalid capture content without inspecting error messages.
type CaptureError struct{ Err error }

func (e *CaptureError) Error() string { return e.Err.Error() }
func (e *CaptureError) Unwrap() error { return e.Err }

// ReadCapture detects cURL or HAR by content, independently of the filename.
func ReadCapture(reader io.Reader) ([]Exchange, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxCaptureBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read capture: %w", err)
	}
	if len(data) > MaxCaptureBytes {
		return nil, &CaptureError{Err: fmt.Errorf("capture exceeds the 32 MiB size limit")}
	}
	data = bytes.TrimSpace(bytes.TrimPrefix(bytes.TrimSpace(data), []byte{0xef, 0xbb, 0xbf}))
	var exchanges []Exchange
	if bytes.HasPrefix(data, []byte("{")) {
		exchanges, err = readHAR(data)
	} else {
		exchanges, err = readCURL(string(data))
	}
	if err != nil {
		return nil, &CaptureError{Err: err}
	}
	return exchanges, nil
}

type harHeader struct{ Name, Value string }
type harBody struct {
	Text, Encoding string
}
type harEntry struct {
	StartedAt time.Time `json:"startedDateTime"`
	Request   struct {
		URL, Method string
		Headers     []harHeader
		PostData    harBody
	}
	Response struct {
		Status  int
		Content harBody
		Headers []harHeader
	}
	Time *float64
}

func readHAR(data []byte) ([]Exchange, error) {
	var archive struct {
		Log *struct{ Entries []harEntry }
	}
	if err := json.Unmarshal(data, &archive); err != nil || archive.Log == nil {
		return nil, fmt.Errorf("invalid HAR: expected an HTTP Archive with log.entries")
	}
	var exchanges []Exchange
	for i, entry := range archive.Log.Entries {
		project, namespace, ok := fetchTarget(entry.Request.URL)
		if !ok || !strings.EqualFold(entry.Request.Method, "POST") {
			continue
		}
		if entry.Response.Status < 0 || entry.Response.Status > 599 {
			return nil, fmt.Errorf("HAR entry %d: invalid HTTP response status", i+1)
		}
		body, err := decodeBody(entry.Request.PostData)
		if err != nil {
			return nil, fmt.Errorf("HAR entry %d: %w", i+1, err)
		}
		facts, err := requestFacts(body)
		if err != nil {
			return nil, fmt.Errorf("HAR entry %d: missing or invalid fetch request body", i+1)
		}
		exchange := Exchange{Entry: i + 1, Project: project, Namespace: namespace, StartedAt: entry.StartedAt, Status: entry.Response.Status, Facts: facts}
		exchange.Format = "har"
		headers := make(http.Header)
		for _, header := range entry.Request.Headers {
			recordCaptureHeader(headers, header.Name, header.Value)
		}
		prepareExchange(&exchange, body, headers)
		captureTime(&exchange, entry)
		if entry.Response.Status == 200 && entry.Response.Content.Text != "" {
			response, decodeErr := decodeBody(entry.Response.Content)
			if decodeErr != nil {
				return nil, fmt.Errorf("HAR entry %d response: %w", i+1, decodeErr)
			}
			readCapturedResponse(&exchange, response)
		}
		exchanges = append(exchanges, exchange)
	}
	if len(exchanges) == 0 {
		return nil, fmt.Errorf("HAR contains no Firebase Remote Config :fetch POST requests")
	}
	return exchanges, nil
}

func decodeBody(body harBody) ([]byte, error) {
	data := []byte(body.Text)
	if body.Encoding != "" {
		if body.Encoding != "base64" {
			return nil, fmt.Errorf("unsupported body encoding")
		}
		decoded, err := base64.StdEncoding.DecodeString(body.Text)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 body")
		}
		data = decoded
	}
	if bytes.HasPrefix(data, []byte{0x1f, 0x8b}) {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("invalid gzip body")
		}
		defer func() { _ = reader.Close() }()
		data, err = io.ReadAll(io.LimitReader(reader, MaxCaptureBytes+1))
		if err != nil || len(data) > MaxCaptureBytes {
			return nil, fmt.Errorf("invalid or oversized gzip body")
		}
	}
	return data, nil
}

func fetchTarget(raw string) (string, string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "firebaseremoteconfig.googleapis.com" || u.User != nil {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 5 || parts[0] != "v1" || parts[1] != "projects" || parts[2] == "" || parts[3] != "namespaces" {
		return "", "", false
	}
	namespace, ok := strings.CutSuffix(parts[4], ":fetch")
	return parts[2], namespace, ok && (namespace == "firebase" || namespace == "firebase-server")
}

func requestFacts(body []byte) (map[string]Fact, error) {
	body, err := normalizeFetchJSON(body)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("expected a JSON fetch request object")
	}
	facts := make(map[string]Fact)
	aliases := map[string][]string{
		"app.id": {"app_id", "appId"}, "app.version": {"app_version", "appVersion"},
		"app.build": {"app_build", "appBuild"}, "app.firebaseInstallationId": {"app_instance_id", "appInstanceId"},
		"device.language": {"language_code", "languageCode"}, "device.country": {"country_code", "countryCode"},
		"app.firstOpenTimestamp": {"first_open_time", "firstOpenTime"},
	}
	for key, names := range aliases {
		for _, name := range names {
			if raw, ok := fields[name]; ok {
				fact, err := scalarFact(raw)
				if err != nil {
					return nil, fmt.Errorf("invalid fetch field %s", name)
				}
				facts[key] = fact
				break
			}
		}
	}
	for _, spec := range []struct {
		prefix string
		names  []string
	}{
		{"app.customSignal", []string{"custom_signals", "customSignals"}},
		{"app.userProperty", []string{"analytics_user_properties", "analyticsUserProperties"}},
	} {
		for _, name := range spec.names {
			if raw, ok := fields[name]; ok {
				var values map[string]json.RawMessage
				if err := json.Unmarshal(raw, &values); err != nil {
					return nil, fmt.Errorf("invalid fetch field %s", name)
				}
				for property, value := range values {
					fact, err := scalarFact(value)
					if err != nil {
						return nil, fmt.Errorf("invalid value in fetch field %s", name)
					}
					facts[spec.prefix+"["+quote(property)+"]"] = fact
				}
			}
		}
	}
	return facts, nil
}

func scalarFact(raw json.RawMessage) (Fact, error) {
	fact := Fact{Known: true, Source: "request"}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		fact.Absent = true
		return fact, nil
	}
	if json.Unmarshal(raw, &fact.Value) == nil {
		return fact, nil
	}
	var number json.Number
	if json.Unmarshal(raw, &number) == nil {
		fact.Value = number.String()
		return fact, nil
	}
	var boolean bool
	if json.Unmarshal(raw, &boolean) == nil {
		fact.Value = fmt.Sprint(boolean)
		return fact, nil
	}
	return Fact{}, fmt.Errorf("expected a scalar value")
}
