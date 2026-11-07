package evaluate

import (
	"strings"
	"testing"
)

const fullChromeUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.6045.123 Safari/537.36"
const reducedChromeUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func TestHeadersSupplyBrowserOSAndLanguageWithoutPrompts(t *testing.T) {
	exchange := capturedCURL(t, `{"app_id":"1:123:web:abc"}`, "User-Agent: "+fullChromeUA, "Accept-Language: fr;q=0.5, pt-BR;q=0.9, en;q=0")
	ctx := &Context{Facts: exchange.Facts, Capture: &exchange, Resolve: func(need Need) (Fact, error) {
		t.Fatalf("prompted despite request header: %s", need.Key)
		return Fact{}, nil
	}}
	for _, expression := range []string{
		`app.browserAndVersion.inOne([browserName('Chrome').version.>=('119')])`,
		`app.operatingSystemAndVersion.inOne([operatingSystemName('Macintosh').version.==('13.6.1')])`,
		`device.language in ['pt-BR']`,
	} {
		result, err := Evaluate(expression, ctx)
		if err != nil || result.State != True {
			t.Fatalf("%s: %#v, %v", expression, result, err)
		}
	}
	exchange = capturedCURL(t, `{"app_id":"1:123:web:abc","language_code":"de"}`, "Accept-Language: en-US")
	if exchange.Facts["device.language"].Value != "de" {
		t.Fatal("header replaced the SDK language")
	}
	entry := harEntry{}
	entry.Request.PostData.Text = `{"app_id":"1:123:web:abc"}`
	entry.Request.Headers = []harHeader{{Name: "USER-AGENT", Value: fullChromeUA}}
	har := capturedHAR(t, entry)
	if har.Facts["app.browserAndVersion.name"].Value != "Chrome" {
		t.Fatal("HAR headers were ignored")
	}
	command := "curl '" + fetchURL + "' -A '" + fullChromeUA + "' --data '{}'"
	exchanges, err := ReadCapture(strings.NewReader(command))
	if err != nil || exchanges[0].Facts["app.browserAndVersion.name"].Value != "Chrome" {
		t.Fatalf("-A headers were ignored: %v", err)
	}
}

func TestReducedVersionsNeedHintsOrExplicitAnswers(t *testing.T) {
	exchange := capturedCURL(t, `{"app_id":"1:123:web:abc"}`, "User-Agent: "+reducedChromeUA)
	ctx := &Context{Facts: exchange.Facts, Capture: &exchange}
	result, _ := Evaluate(`app.browserAndVersion.inOne([browserName('Chrome').anyVersion])`, ctx)
	if result.State != True {
		t.Fatal("reduction hid the known browser name")
	}
	result, _ = Evaluate(`app.browserAndVersion.inOne([browserName('Chrome').version.>=('120')])`, ctx)
	if result.State != Unknown {
		t.Fatal("frozen browser version treated as exact")
	}
	exchange = capturedCURL(t, `{"app_id":"1:123:web:abc"}`, "User-Agent: "+reducedChromeUA, `Sec-CH-UA-Full-Version-List: "Not,A;Brand";v="99.0.0.0", "Google Chrome";v="120.0.6099.234"`, `Sec-CH-UA-Platform: "macOS"`, `Sec-CH-UA-Platform-Version: "14.2.1"`)
	ctx = &Context{Facts: exchange.Facts, Capture: &exchange}
	for _, expression := range []string{
		`app.browserAndVersion.inOne([browserName('Chrome').version.==('120.0.6099.234')])`,
		`app.operatingSystemAndVersion.inOne([operatingSystemName('Macintosh').version.==('14.2.1')])`,
	} {
		result, _ = Evaluate(expression, ctx)
		if result.State != True {
			t.Fatalf("full hints were ignored: %#v", result)
		}
	}
	ctx.Facts["app.browserAndVersion.version"] = Fact{Known: true, Value: "121", Source: "command line"}
	result, _ = Evaluate(`app.browserAndVersion.inOne([browserName('Chrome').version.==('121')])`, ctx)
	if result.State != True {
		t.Fatal("explicit version did not override derived headers")
	}
}

func TestNativeAppsDoNotUseWebOnlyHeaderFacts(t *testing.T) {
	exchange := capturedCURL(t, `{"app_id":"1:123:ios:abc"}`, "User-Agent: "+fullChromeUA)
	result, _ := Evaluate(`app.browserAndVersion.inOne([browserName('Chrome').anyVersion])`, &Context{Facts: exchange.Facts, Capture: &exchange})
	if result.State != False {
		t.Fatalf("native app acquired a web browser: %#v", result)
	}
}

func TestInvalidAndConflictingHeadersDoNotBecomeFacts(t *testing.T) {
	exchange := capturedCURL(t, `{}`, "User-Agent: "+fullChromeUA, "User-Agent: unknown conflicting agent", "Authorization: secret-token", "Cookie: secret-cookie", "Accept-Language: *;q=1,fr;q=broken,en;q=0")
	if exchange.Facts["app.browserAndVersion.name"].Known || exchange.Facts["device.language"].Known {
		t.Fatal("ambiguous or unacceptable headers became known facts")
	}
	for _, fact := range exchange.Facts {
		if strings.Contains(fact.Value, "secret") {
			t.Fatal("credential retained")
		}
	}
	exchange = capturedCURL(t, `{}`, "User-Agent: "+fullChromeUA, `Sec-CH-UA: "Microsoft Edge";v="119"`, `Sec-CH-UA-Platform: "Android"`, `Sec-CH-UA-Platform-Version: "14"`, "Accept-Language: fr;q=NaN")
	if exchange.Facts["app.browserAndVersion.name"].Known || exchange.Facts["app.operatingSystemAndVersion.name"].Known || exchange.Facts["app.operatingSystemAndVersion.version"].Known || exchange.Facts["device.language"].Known {
		t.Fatal("conflicting header types or invalid quality became confirmed facts")
	}
}
