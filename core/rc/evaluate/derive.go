package evaluate

import (
	"strings"

	"github.com/yumauri/fbrcm/core"
)

func (c *Context) derive(need Need) (Fact, bool) {
	platform, validPlatform := platformFromFacts(c.Facts)
	if need.Key == "device.os" {
		app := c.Facts["app.id"]
		if validPlatform {
			return Fact{Known: true, Value: string(platform), Source: app.Source + " app ID", DerivedFrom: []string{"app.id"}}, true
		}
	}
	if c.Capture != nil {
		if validPlatform && platform != core.AppPlatformWeb && (strings.HasPrefix(need.Key, "app.browserAndVersion.") || strings.HasPrefix(need.Key, "app.operatingSystemAndVersion.")) {
			return Fact{Known: true, Absent: true, Source: "request app ID", DerivedFrom: []string{"app.id"}}, true
		}
		if c.Capture.Complete["app.customSignal"] && (need.Key == "app.version" || need.Key == "app.build" || need.Key == "app.firebaseInstallationId") {
			return Fact{Known: true, Absent: true, Source: "request"}, true
		}
		for prefix, complete := range c.Capture.Complete {
			if complete && strings.HasPrefix(need.Key, prefix+"[") {
				return Fact{Known: true, Absent: true, Source: "request"}, true
			}
		}
	}
	return Fact{}, false
}

func platformFromFacts(facts map[string]Fact) (core.AppPlatform, bool) {
	app := facts["app.id"]
	id, valid := core.ParseFirebaseAppID(app.Value)
	return id.Platform, valid && app.Known && !app.Absent
}
