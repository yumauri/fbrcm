package mcpserver

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/cli/contract"
)

func TestArgvKeepsPolicyAndFlagLikeValuesSeparate(t *testing.T) {
	c := contract.Capability{Path: []string{"get"}, Arguments: []contract.ArgumentCapability{{Name: "parameter"}}, Flags: []contract.FlagCapability{{Name: "--filter", Type: "stringArray"}}}
	in := Invocation{Arguments: map[string]json.RawMessage{"parameter": json.RawMessage(`"--profile=other"`)}, Options: map[string]json.RawMessage{"filter": json.RawMessage(`["a,b","--stateless=false"]`)}}
	argv, err := in.Argv(c, Options{Stateless: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	if argv[len(argv)-2] != "--" || argv[len(argv)-1] != "--profile=other" || !slices.Contains(argv, "--filter=a,b") || !slices.Contains(argv, "--filter=--stateless=false") {
		t.Fatalf("unsafe argv: %v", argv)
	}
	in.Options["yes"] = json.RawMessage(`true`)
	if _, err := in.Argv(c, Options{Stateless: true}, false); err == nil {
		t.Fatal("model supplied a bound option")
	}
}

func TestLaunchOptionsValidateBoundaries(t *testing.T) {
	base := Options{Toolsets: []string{"inspect"}, Confirmation: "host", BrowserAuth: "auto", RequestTimeout: time.Minute, AuthTimeout: time.Minute}
	if err := base.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Options){
		func(o *Options) { o.Stateless = true; o.Profile = "work" }, func(o *Options) { o.Stateless = true; o.AllowHooks = true },
		func(o *Options) { o.Confirmation = "ask" }, func(o *Options) { o.BrowserAuth = "always" }, func(o *Options) { o.Toolsets = nil },
		func(o *Options) { o.Toolsets = []string{"unknown"} }, func(o *Options) { o.RequestTimeout = 0 }, func(o *Options) { o.AuthTimeout = -time.Second },
	} {
		o := base
		change(&o)
		if err := o.Validate(); err == nil {
			t.Fatalf("accepted invalid options: %#v", o)
		}
	}
}
