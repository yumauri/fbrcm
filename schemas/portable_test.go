package schemas

import (
	"encoding/json"
	"reflect"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

var portabilityCases = map[string]string{
	"unrestricted":          `{"properties":{"value":true},"required":["value"]}`,
	"forbidden":             `{"properties":{"value":false}}`,
	"empty":                 `{}`,
	"nullable":              `{"type":["string","null"],"minLength":2}`,
	"overlapping_types":     `{"type":["integer","number"],"minimum":1}`,
	"existing_compositions": `{"type":["number","null"],"anyOf":[{"const":1},{"type":"string"},{"type":"null"}],"allOf":[{"not":{"const":2}}]}`,
	"conditional_false":     `{"if":{"properties":{"value":{"const":null}},"required":["value"]},"then":false,"else":{"type":"object"}}`,
	"conditional_true":      `{"if":{"type":"object"},"then":true,"else":{"type":["string","null"]}}`,
	"negation":              `{"anyOf":[{"not":true},{"not":false}]}`,
	"tuple":                 `{"type":"array","prefixItems":[true,false],"items":false}`,
	"dependent_schema":      `{"dependentSchemas":{"value":false},"additionalProperties":true}`,
	"pattern_properties":    `{"patternProperties":{"^v":false},"additionalProperties":true}`,
	"contains":              `{"type":"array","contains":true,"minContains":1,"maxContains":2}`,
	"property_names":        `{"propertyNames":false}`,
	"unevaluated":           `{"type":["object","null"],"anyOf":[{"properties":{"value":true}},{"properties":{"other":true}}],"unevaluatedProperties":false}`,
	"reference_library":     `{"$defs":{"library":{"$defs":{"value":{"type":["string","null"]}}}},"$ref":"#/$defs/library/$defs/value"}`,
	"annotated_any_json":    `{"title":"Free-form value","description":"Accept any JSON","default":false,"x-fbrcm-kind":"json"}`,
	"content_schema":        `{"type":"string","contentMediaType":"application/json","contentSchema":true}`,
}

func decodeSchema(t testing.TB, raw string) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

func compileSchema(t testing.TB, schema map[string]any) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	const id = "https://fbrcm.invalid/portability-test"
	if err := compiler.AddResource(id, schema); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile(id)
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func TestMakePortablePreservesValidation(t *testing.T) {
	values := []any{nil, false, true, float64(-1), float64(0), float64(1), float64(1.5), float64(2), "", "a", "ab", []any{}, []any{nil}, []any{true, false}, []any{1.0, 2.0, 3.0}, map[string]any{}}
	for _, value := range append([]any(nil), values...) {
		values = append(values, map[string]any{"value": value}, map[string]any{"other": value}, map[string]any{"value": value, "other": nil})
	}
	for name, raw := range portabilityCases {
		t.Run(name, func(t *testing.T) {
			original := compileSchema(t, decodeSchema(t, raw))
			portable := decodeSchema(t, raw)
			MakePortable(portable)
			converted := compileSchema(t, portable)
			for _, value := range values {
				before, after := original.Validate(value), converted.Validate(value)
				if (before == nil) != (after == nil) {
					t.Fatalf("validation changed for %#v: before=%v after=%v", value, before, after)
				}
			}
			first, _ := json.Marshal(portable)
			MakePortable(portable)
			second, _ := json.Marshal(portable)
			if string(first) != string(second) {
				t.Fatal("normalization is not idempotent")
			}
		})
	}
}

func TestMakePortableUsesExplicitEquivalentForms(t *testing.T) {
	schema := decodeSchema(t, `{"properties":{"forbidden":false,"nullable":{"type":["string","null"],"minLength":2},"anything":true,"empty":{}},"additionalProperties":false}`)
	MakePortable(schema)
	properties := schema["properties"].(map[string]any)
	if !reflect.DeepEqual(properties["forbidden"], map[string]any{"not": map[string]any{}}) {
		t.Fatalf("forbidden property changed: %v", properties["forbidden"])
	}
	want := map[string]any{"anyOf": []any{map[string]any{"type": "string"}, map[string]any{"type": "null"}}, "minLength": float64(2)}
	if !reflect.DeepEqual(properties["nullable"], want) {
		t.Fatalf("nullable constraint changed: %v", properties["nullable"])
	}
	for _, name := range []string{"anything", "empty"} {
		if !reflect.DeepEqual(properties[name], map[string]any{"anyOf": jsonValueTypes()}) {
			t.Fatalf("%s is not explicitly any JSON: %v", name, properties[name])
		}
	}
	if schema["additionalProperties"] != false {
		t.Fatal("boolean-valued keyword changed")
	}
}

func TestMakePortableDoesNotRewriteInstanceData(t *testing.T) {
	const raw = `{"type":"object","properties":{"type":{"type":["string","null"]}},"const":{"properties":{"value":false},"type":["string","null"]},"enum":[true,false,{}],"default":{"items":false},"examples":[{"properties":{"value":true}}],"x-fbrcm-test":{"type":["string","null"],"items":false}}`
	original, portable := decodeSchema(t, raw), decodeSchema(t, raw)
	MakePortable(portable)
	for _, key := range []string{"const", "enum", "default", "examples", "x-fbrcm-test"} {
		if !reflect.DeepEqual(original[key], portable[key]) {
			t.Errorf("instance data changed under %s", key)
		}
	}
}

func FuzzMakePortablePreservesValidation(f *testing.F) {
	type pair struct{ before, after *jsonschema.Schema }
	compiled := make(map[string]pair, len(portabilityCases))
	for name, raw := range portabilityCases {
		before := compileSchema(f, decodeSchema(f, raw))
		portable := decodeSchema(f, raw)
		MakePortable(portable)
		compiled[name] = pair{before, compileSchema(f, portable)}
	}
	for _, value := range []string{`null`, `false`, `1`, `1.5`, `"ab"`, `[]`, `[1,null]`, `{}`, `{"value":null}`, `{"value":{},"other":true}`} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			t.Skip()
		}
		for name, p := range compiled {
			if (p.before.Validate(value) == nil) != (p.after.Validate(value) == nil) {
				t.Fatalf("%s validation changed for %s", name, raw)
			}
		}
	})
}
