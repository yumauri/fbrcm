package contract

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

type responseSchemaNested struct {
	Enabled bool `json:"enabled"`
}

type responseSchemaItem struct {
	Name     string                `json:"name"`
	Optional *string               `json:"optional,omitempty"`
	Tags     []string              `json:"tags"`
	Nested   *responseSchemaNested `json:"nested"`
}

func TestResponseDataSchemaDescribesCollectionDTO(t *testing.T) {
	cmd := &cobra.Command{Use: "example"}
	RegisterResponse(cmd, []responseSchemaItem{})

	schema, err := ResponseDataSchema(cmd)
	if err != nil {
		t.Fatal(err)
	}
	variants := schema["oneOf"].([]any)
	collection := variants[0].(map[string]any)
	if collection["additionalProperties"] != false {
		t.Fatalf("collection schema is not strict: %#v", collection)
	}
	properties := collection["properties"].(map[string]any)
	items := properties["items"].(map[string]any)["items"].(map[string]any)
	if items["additionalProperties"] != false {
		t.Fatalf("item schema is not strict: %#v", items)
	}
	itemProperties := items["properties"].(map[string]any)
	for _, name := range []string{"name", "optional", "tags", "nested"} {
		if _, ok := itemProperties[name]; !ok {
			t.Errorf("item schema is missing %q", name)
		}
	}
	optional := itemProperties["optional"].(map[string]any)
	if got := optional["type"]; !jsonEqual(got, []string{"string", "null"}) {
		t.Fatalf("optional type = %#v", got)
	}
	required := items["required"].([]string)
	if containsString(required, "optional") {
		t.Fatalf("omitempty field is required: %#v", required)
	}
	if variants[1].(map[string]any)["type"] != "null" {
		t.Fatalf("failure data variant = %#v", variants[1])
	}
}

func TestResponseDataSchemaRequiresRegistration(t *testing.T) {
	if _, err := ResponseDataSchema(&cobra.Command{Use: "missing"}); err == nil {
		t.Fatal("unregistered command returned a response schema")
	}
}

func TestNoDataResponseSchemaIsNull(t *testing.T) {
	cmd := &cobra.Command{Use: "no-data"}
	RegisterNoData(cmd)
	schema, err := ResponseDataSchema(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if schema["type"] != "null" {
		t.Fatalf("schema = %#v", schema)
	}
}

func TestSuccessDataSchemaNeverAcceptsNull(t *testing.T) {
	cmd := &cobra.Command{Use: "example"}
	RegisterResponse(cmd, responseSchemaItem{})
	schema, err := ResponseSuccessDataSchema(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if schema["type"] == "null" {
		t.Fatalf("success schema = %#v", schema)
	}

	noData := &cobra.Command{Use: "no-data"}
	RegisterNoData(noData)
	schema, err = ResponseSuccessDataSchema(noData)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := schema["not"]; !ok {
		t.Fatalf("no-data success schema = %#v, want unsatisfiable schema", schema)
	}
}

func jsonEqual(left, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}

func containsString(values []string, target string) bool {
	return slices.Contains(values, target)
}
