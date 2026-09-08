package ops

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/ops/contract"
)

func TestStructuredOptionsPreserveStringsAndArrays(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.StringArray("array", []string{"default"}, "")
	flags.StringSlice("slice", []string{"default"}, "")
	flags.String("value", "", "")
	flags.Bool("switch", true, "")
	for _, name := range []string{"array", "slice"} {
		if err := bindOption(flags, name, json.RawMessage(`["a,b","--profile=other","quoted\"value"]`)); err != nil {
			t.Fatal(err)
		}
		want := []string{"a,b", "--profile=other", "quoted\"value"}
		got := flags.Lookup(name).Value.(pflag.SliceValue).GetSlice()
		if !reflect.DeepEqual(got, want) || !flags.Changed(name) {
			t.Fatalf("%s = %#v, changed=%t", name, got, flags.Changed(name))
		}
	}
	if err := bindOption(flags, "value", json.RawMessage(`"--stateless=false"`)); err != nil {
		t.Fatal(err)
	}
	if err := bindOption(flags, "switch", json.RawMessage(`false`)); err != nil {
		t.Fatal(err)
	}
	if flags.Lookup("value").Value.String() != "--stateless=false" || flags.Lookup("switch").Value.String() != "false" {
		t.Fatal("structured value was reinterpreted")
	}
}

func TestPositionalsAllowsRequiredArgumentAfterOmittedOptionalArgument(t *testing.T) {
	input := Input{Arguments: map[string]json.RawMessage{"name": json.RawMessage(`"Beta users"`)}}
	capability := contract.Capability{Arguments: []contract.ArgumentCapability{
		{Name: "project"},
		{Name: "name", Required: true},
	}}

	got, err := input.Positionals(capability)
	if err != nil {
		t.Fatalf("Positionals returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "Beta users" {
		t.Fatalf("Positionals = %#v, want compact name argument", got)
	}
}

func TestPositionalsRejectsOptionalArgumentAfterOmittedOptionalArgument(t *testing.T) {
	input := Input{Arguments: map[string]json.RawMessage{"condition": json.RawMessage(`"Beta users"`)}}
	capability := contract.Capability{Arguments: []contract.ArgumentCapability{
		{Name: "project"},
		{Name: "condition"},
	}}

	if _, err := input.Positionals(capability); err == nil {
		t.Fatal("Positionals accepted an optional argument after an omitted optional argument")
	}
}
