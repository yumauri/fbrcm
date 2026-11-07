package evaluate

import (
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ua-parser/uap-go/uaparser"
)

var requestUAParser = sync.OnceValues(func() (*uaparser.Parser, error) { return uaparser.New() })
var clientHintBrand = regexp.MustCompile(`"((?:\\.|[^"\\])*)"\s*;\s*v\s*=\s*"([0-9.]+)"`)

func headerFacts(facts map[string]Fact, headers http.Header) {
	for name, values := range headers {
		if name == "Accept-Language" {
			headers.Set(name, strings.Join(values, ","))
			continue
		}
		for _, value := range values {
			if value != values[0] {
				headers.Set(name, "")
				break
			}
		}
	}
	set := func(key, value, source string) {
		if _, exists := facts[key]; !exists && value != "" && value != "Other" {
			facts[key] = Fact{Known: true, Value: value, Source: source}
		}
	}
	setHintName := func(prefix, value string) bool {
		previous := facts[prefix+".name"]
		if previous.Reason != "" || previous.Known && previous.Value != value {
			reason := "Request headers disagree about " + prefix + ", so the Firebase interpretation needs verification"
			facts[prefix+".name"] = Fact{Reason: reason, Source: "Client Hints header"}
			facts[prefix+".version"] = Fact{Reason: reason, Source: "Client Hints header"}
			return false
		}
		facts[prefix+".name"] = Fact{Known: true, Value: value, Source: "Client Hints header"}
		return true
	}
	ua := headers.Get("User-Agent")
	if len(ua) > 0 && len(ua) <= 8192 {
		if parser, err := requestUAParser(); err == nil {
			client := parser.Parse(ua)
			browser := client.UserAgent.Family
			switch browser {
			case "Chrome Mobile", "Chrome Mobile iOS", "Chrome WebView":
				browser = "Chrome"
			case "Firefox Mobile", "Firefox iOS":
				browser = "Firefox"
			case "Mobile Safari", "Mobile Safari UI/WKWebView":
				browser = "Safari"
			case "Edge Mobile":
				browser = "Edge"
			case "Opera Mobile", "Opera Mini":
				browser = "Opera"
			}
			os := client.Os.Family
			if os == "Mac OS X" {
				os = "Macintosh"
			}
			set("app.browserAndVersion.name", browser, "User-Agent header")
			set("app.browserAndVersion.version", client.UserAgent.ToVersionString(), "User-Agent header")
			set("app.operatingSystemAndVersion.name", os, "User-Agent header")
			set("app.operatingSystemAndVersion.version", client.Os.ToVersionString(), "User-Agent header")
			if strings.Contains(ua, ".0.0.0") {
				facts["app.browserAndVersion.version"] = Fact{Reason: "The User-Agent reports a reduced browser version, so the full version is unavailable", Source: "User-Agent header"}
			}
			// Chromium freezes these OS versions. Do not treat them as the
			// actual version when a version-sensitive rule needs more detail.
			if strings.Contains(ua, "Chrome/") && (strings.Contains(ua, "Windows NT 10.0") || strings.Contains(ua, "Mac OS X 10_15_7") || strings.Contains(ua, "Android 10; K")) {
				facts["app.operatingSystemAndVersion.version"] = Fact{Reason: "The User-Agent reports a frozen operating-system version, so the actual version is unavailable", Source: "User-Agent header"}
			}
		}
	}
	// Structured Client Hints can provide the versions hidden by UA reduction.
	brands := headers.Get("Sec-CH-UA-Full-Version-List")
	full := brands != ""
	if !full {
		brands = headers.Get("Sec-CH-UA")
	}
	for _, brand := range clientHintBrand.FindAllStringSubmatch(brands, -1) {
		name, err := strconv.Unquote(`"` + brand[1] + `"`)
		if err != nil {
			continue
		}
		browser := map[string]string{"Google Chrome": "Chrome", "Microsoft Edge": "Edge", "Opera": "Opera"}[name]
		if browser == "" {
			continue
		}
		if !setHintName("app.browserAndVersion", browser) {
			continue
		}
		if full {
			facts["app.browserAndVersion.version"] = Fact{Known: true, Value: brand[2], Source: "Client Hints header"}
		}
	}
	if platform, err := strconv.Unquote(headers.Get("Sec-CH-UA-Platform")); err == nil {
		if platform == "macOS" {
			platform = "Macintosh"
		}
		consistent := setHintName("app.operatingSystemAndVersion", platform)
		if version, err := strconv.Unquote(headers.Get("Sec-CH-UA-Platform-Version")); err == nil && version != "" {
			// Windows Client Hint versions are platform API versions, not the
			// Windows versions used in Remote Config targets.
			if consistent && platform != "Windows" {
				facts["app.operatingSystemAndVersion.version"] = Fact{Known: true, Value: version, Source: "Client Hints header"}
			}
		}
	}
	if _, present := facts["device.language"]; !present {
		if language := preferredLanguage(headers.Get("Accept-Language")); language != "" {
			set("device.language", language, "Accept-Language header")
		}
	}
}

func preferredLanguage(header string) string {
	type preference struct {
		language string
		quality  float64
	}
	var choices []preference
	for value := range strings.SplitSeq(header, ",") {
		language, parameters, _ := strings.Cut(strings.TrimSpace(value), ";")
		if language == "" || language == "*" {
			continue
		}
		quality := 1.0
		if parameters != "" {
			name, weight, ok := strings.Cut(strings.TrimSpace(parameters), "=")
			if !ok || name != "q" {
				continue
			}
			var err error
			quality, err = finiteNumber(weight)
			if err != nil || quality <= 0 || quality > 1 {
				continue
			}
		}
		choices = append(choices, preference{language, quality})
	}
	sort.SliceStable(choices, func(i, j int) bool { return choices[i].quality > choices[j].quality })
	if len(choices) == 0 {
		return ""
	}
	return choices[0].language
}
