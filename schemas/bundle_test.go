package schemas

import (
	"encoding/json"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

func TestBundleIsDeterministicAndResolvesWithoutExternalLoader(t *testing.T) {
	for _, id := range []string{"urn:fbrcm:schema:cli:1.0.0:command:get:input", "urn:fbrcm:schema:cli:1.0.0:command:get:response", "urn:fbrcm:schema:cli:1.0.0:command:apply:input"} {
		a, err := Bundle(id)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Bundle(id)
		if err != nil {
			t.Fatal(err)
		}
		aJSON, _ := json.Marshal(a)
		bJSON, _ := json.Marshal(b)
		if string(aJSON) != string(bJSON) {
			t.Fatal("nondeterministic bundle")
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource("https://fbrcm.invalid/test", a); err != nil {
			t.Fatal(err)
		}
		if _, err := compiler.Compile("https://fbrcm.invalid/test"); err != nil {
			t.Fatal(err)
		}
	}
}
