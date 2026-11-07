package evaluate

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core"
)

func recordCaptureHeader(headers http.Header, name, value string) {
	if len(value) > 8192 {
		return
	}
	name = http.CanonicalHeaderKey(strings.TrimSpace(name))
	switch name {
	case "User-Agent", "Accept-Language", "Sec-Ch-Ua", "Sec-Ch-Ua-Full-Version-List", "Sec-Ch-Ua-Platform", "Sec-Ch-Ua-Platform-Version", "Sec-Ch-Ua-Mobile":
		headers.Add(name, strings.TrimSpace(value))
	}
}

func prepareExchange(exchange *Exchange, body []byte, headers http.Header) {
	exchange.Complete = make(map[string]bool)
	standard, _ := normalizeFetchJSON(body) // Already validated by requestFacts.
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(standard, &fields)
	platform, validPlatform := platformFromFacts(exchange.Facts)
	exchange.Complete["app.customSignal"] = validPlatform
	_, snakeProperties := fields["analytics_user_properties"]
	_, camelProperties := fields["analyticsUserProperties"]
	exchange.Complete["app.userProperty"] = (snakeProperties || camelProperties) && validPlatform && platform != core.AppPlatformWeb
	for _, key := range []string{"time_zone", "timeZone"} {
		if raw, ok := fields[key]; ok {
			_ = json.Unmarshal(raw, &exchange.TimeZone)
		}
	}
	if country, ok := exchange.Facts["device.country"]; ok && country.Known && !country.Absent {
		country.Known, country.Source = false, "request locale"
		country.Reason = "The request reports locale country " + quote(country.Value) + ", but Firebase may use IP or Analytics country instead"
		exchange.Facts["device.country"] = country
	}
	if _, ok := exchange.Facts["app.firstOpenTimestamp"]; !ok {
		if fot := exchange.Facts["app.userProperty['_fot']"]; fot.Known && !fot.Absent {
			if millis, err := strconv.ParseInt(fot.Value, 10, 64); err == nil && millis > 0 {
				firstOpen := time.UnixMilli(millis).UTC()
				if platform == core.AppPlatformIOS {
					firstOpen = firstOpen.Truncate(time.Second)
				}
				exchange.Facts["app.firstOpenTimestamp"] = Fact{Known: true, Value: firstOpen.Format(time.RFC3339Nano), Source: "request Analytics _fot", DerivedFrom: []string{"app.userProperty['_fot']"}}
			}
		}
	}
	headerFacts(exchange.Facts, headers)
	if validPlatform && platform != core.AppPlatformWeb {
		for key := range exchange.Facts {
			if strings.HasPrefix(key, "app.browserAndVersion.") || strings.HasPrefix(key, "app.operatingSystemAndVersion.") {
				delete(exchange.Facts, key)
			}
		}
	}
}

func captureTime(exchange *Exchange, entry harEntry) {
	var server time.Time
	for _, header := range entry.Response.Headers {
		if strings.EqualFold(header.Name, "Date") {
			server, _ = http.ParseTime(header.Value)
		}
	}
	if !server.IsZero() {
		fact := Fact{Known: true, Value: server.Format(time.RFC3339Nano), Source: "response Date header", TimeMax: server.Add(time.Second)}
		if entry.Time != nil && *entry.Time >= 0 && *entry.Time < 86400000 {
			fact.TimeMin = server.Add(-time.Duration(*entry.Time * float64(time.Millisecond)))
		}
		exchange.Facts["device.dateTime"] = fact
	} else if !entry.StartedAt.IsZero() {
		fact := Fact{Known: true, Value: entry.StartedAt.Format(time.RFC3339Nano), Source: "capture timestamp", TimeMin: entry.StartedAt}
		if entry.Time != nil && *entry.Time >= 0 && *entry.Time < 86400000 {
			fact.TimeMax = entry.StartedAt.Add(time.Duration(*entry.Time * float64(time.Millisecond)))
		}
		exchange.Facts["device.dateTime"] = fact
	}
}
