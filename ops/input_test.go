package ops

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/spf13/pflag"
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
