package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type payload struct {
	Request struct {
		Method      *string             `json:"method"`
		Destination *string             `json:"destination"`
		Path        *string             `json:"path"`
		Query       *string             `json:"query"`
		Headers     map[string][]string `json:"headers"`
	} `json:"request"`
}

type allowedRequest struct {
	Method         string `json:"method"`
	Host           string `json:"host"`
	Path           string `json:"path"`
	Query          string `json:"query,omitempty"`
	QuotaProjectID string `json:"quota_project_id,omitempty"`
}

func main() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail("read middleware payload: %v", err)
	}
	var value payload
	if err := json.Unmarshal(raw, &value); err != nil {
		fail("decode middleware payload: %v", err)
	}
	allowed, err := decodeAllowedRequests(os.Getenv("FBRCM_E2E_ALLOWED_REQUESTS"))
	if err != nil {
		fail("decode allowed requests: %v", err)
	}
	if err := validate(value, allowed); err != nil {
		fail("%v", err)
	}
	if _, err := os.Stdout.Write(raw); err != nil {
		fail("write middleware payload: %v", err)
	}
}

func validate(value payload, allowed []allowedRequest) error {
	method := pointerValue(value.Request.Method)
	destination := pointerValue(value.Request.Destination)
	path := pointerValue(value.Request.Path)
	query := pointerValue(value.Request.Query)
	for _, request := range allowed {
		if method == request.Method &&
			(destination == request.Host || destination == request.Host+":443") &&
			path == request.Path &&
			(request.Query == "" || query == request.Query) &&
			headerValue(value.Request.Headers, "X-Goog-User-Project") == request.QuotaProjectID {
			return nil
		}
	}
	return fmt.Errorf("E2E proxy blocked unexpected request %s %s%s", method, destination, path)
}

func headerValue(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
	}
	return ""
}

func decodeAllowedRequests(raw string) ([]allowedRequest, error) {
	var allowed []allowedRequest
	if err := json.Unmarshal([]byte(raw), &allowed); err != nil {
		return nil, err
	}
	for index := range allowed {
		allowed[index].Method = strings.ToUpper(strings.TrimSpace(allowed[index].Method))
		allowed[index].Host = strings.TrimSpace(allowed[index].Host)
		allowed[index].Path = strings.TrimSpace(allowed[index].Path)
		if allowed[index].Method == "" || allowed[index].Host == "" || allowed[index].Path == "" {
			return nil, fmt.Errorf("entry %d requires method, host, and path", index+1)
		}
	}
	return allowed, nil
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func fail(format string, values ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(42)
}
