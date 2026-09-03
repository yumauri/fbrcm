package mcpserver

import (
	"fmt"
	"maps"
	"reflect"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/yumauri/fbrcm/ops"
)

func compileInput(t *testing.T, input map[string]any) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	const id = "https://fbrcm.invalid/defaults-test"
	if err := compiler.AddResource(id, input); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile(id)
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func TestInputDefaultsPreserveCatalogValidation(t *testing.T) {
	registry, err := ops.NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	arguments := []any{
		map[string]any{}, nil, "invalid",
		map[string]any{"parameter": "feature"},
		map[string]any{"project": "=demo"},
		map[string]any{"project": "^fuzzy"},
		map[string]any{"plan": "-"},
		map[string]any{"projects": []any{"=demo"}},
	}
	options := []any{
		map[string]any{}, nil,
		map[string]any{"all": true},
		map[string]any{"type": "string", "value": "hello"},
		map[string]any{"project": []any{"=demo"}},
		map[string]any{"project": []any{""}},
		map[string]any{"filter": []any{"feature"}},
		map[string]any{"update": true},
		map[string]any{"draft": true},
		map[string]any{"yes": true},
	}
	stdin := []any{nil, map[string]any{"parameters": map[string]any{}}, "invalid"}
	for _, stateless := range []bool{false, true} {
		for _, writes := range []bool{false, true} {
			policy := Options{Stateless: stateless, AllowWrites: writes, Toolsets: []string{"inspect", "edit", "drafts", "plans", "publish", "diagnostics"}}
			for _, capability := range registry.Capabilities() {
				if !policy.allows(capability) {
					continue
				}
				t.Run(fmt.Sprintf("stateless=%t/writes=%t/%s", stateless, writes, capability.ID), func(t *testing.T) {
					input, tool, err := toolInputSchema(capability, policy)
					if err != nil {
						t.Fatal(err)
					}
					published := compileInput(t, input)
					for _, name := range wrapperNames {
						property := input["properties"].(map[string]any)[name].(map[string]any)
						value, advertised := property["default"]
						expected, normalized := tool.defaults[name]
						if advertised != normalized || !reflect.DeepEqual(value, expected) {
							t.Fatalf("%s default does not match normalization", name)
						}
						if advertised && published.Properties[name].Validate(value) != nil {
							t.Fatalf("%s default violates its schema", name)
						}
					}
					for _, args := range arguments {
						for _, opts := range options {
							for _, doc := range stdin {
								for omitted := range 8 {
									value := map[string]any{"arguments": args, "options": opts, "stdin": doc}
									for i, name := range wrapperNames {
										if omitted&(1<<i) != 0 {
											delete(value, name)
										}
									}
									normalized := maps.Clone(value)
									tool.defaults.normalize(normalized)
									want := tool.schema.Validate(normalized) == nil
									if got := published.Validate(value) == nil; got != want {
										t.Fatalf("published valid=%t, normalized valid=%t for %#v", got, want, value)
									}
								}
							}
						}
					}
				})
			}
		}
	}
}

func TestNormalizeOnlyMissingWrappersAndIsolatesDefaults(t *testing.T) {
	defaults := inputDefaults{"arguments": map[string]any{}, "options": map[string]any{}, "stdin": nil}
	first, second := map[string]any{}, map[string]any{}
	defaults.normalize(first)
	first["options"].(map[string]any)["project"] = "must not leak"
	defaults.normalize(second)
	if len(second["options"].(map[string]any)) != 0 {
		t.Fatal("default object shared between invocations")
	}
	explicit := map[string]any{"arguments": nil, "options": "invalid", "stdin": map[string]any{"unknown": true}}
	want := maps.Clone(explicit)
	defaults.normalize(explicit)
	if !reflect.DeepEqual(explicit, want) {
		t.Fatal("explicit values were changed")
	}
}

func TestInputDefaultsRejectUnsupportedRootConstraints(t *testing.T) {
	input := map[string]any{
		"type": "object", "minProperties": 3,
		"properties": map[string]any{
			"arguments": map[string]any{"type": "object"},
			"options":   map[string]any{"type": "object"},
			"stdin":     map[string]any{"type": "null"},
		},
	}
	if _, err := addInputDefaults(input, compileInput(t, input)); err == nil {
		t.Fatal("silently changed an unsupported presence-sensitive constraint")
	}
}
