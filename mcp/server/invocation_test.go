package mcpserver

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/ops/contract"
)

func TestStructuredInputKeepsPolicyAndFlagLikeValuesSeparate(t *testing.T) {
	c := contract.Capability{Path: []string{"get"}, Arguments: []contract.ArgumentCapability{{Name: "parameter"}}, Flags: []contract.FlagCapability{{Name: "--filter", Type: "stringArray"}}}
	in := Invocation{Arguments: map[string]json.RawMessage{"parameter": json.RawMessage(`"--profile=other"`)}, Options: map[string]json.RawMessage{"filter": json.RawMessage(`["a,b","--stateless=false"]`)}}
	args, err := in.Positionals(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "--profile=other" {
		t.Fatalf("changed positional value: %v", args)
	}
	if err := in.Validate(c); err != nil {
		t.Fatal(err)
	}
	in.Options["yes"] = json.RawMessage(`true`)
	if err := in.Validate(c); err == nil {
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
