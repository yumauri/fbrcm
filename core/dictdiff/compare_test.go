package dictdiff

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompareNormalizesTypesAndOmitsEqualProperties(t *testing.T) {
	result, err := Compare(Input{
		EntityName: "demo",
		Left: NamedDictionary{Name: "left", Properties: Dictionary{
			"equal":   String("same"),
			"boolean": Boolean(true),
			"number":  Number(json.Number("1.50")),
			"json":    JSON(`{"enabled":true,"items":[1,2]}`),
			"null":    Null(),
		}},
		Right: NamedDictionary{Name: "right", Properties: Dictionary{
			"equal":   String("same"),
			"boolean": Boolean(false),
			"number":  Number(json.Number("2")),
			"json":    JSON(`{"enabled":false,"items":[1,2]}`),
			"null":    String("null"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Properties) != 4 {
		t.Fatalf("changed properties = %d, want 4", len(result.Properties))
	}
	jsonProperty := propertyByName(t, result, "json")
	if !strings.Contains(jsonProperty.Left.Text, "\n  \"enabled\": true") ||
		!strings.Contains(jsonProperty.Right.Text, "\n  \"enabled\": false") {
		t.Fatalf("JSON was not prettified:\nleft=%s\nright=%s", jsonProperty.Left.Text, jsonProperty.Right.Text)
	}
	nullProperty := propertyByName(t, result, "null")
	if nullProperty.Left.Type != ValueNull || nullProperty.Right.Type != ValueString {
		t.Fatalf("explicit null and string null lost their types: %#v", nullProperty)
	}
	if len(nullProperty.Chunks) != 1 || len(nullProperty.Chunks[0].Rows) != 1 {
		t.Fatalf("same-text values with different types have no visible diff: %#v", nullProperty.Chunks)
	}
}

func TestCompareDistinguishesMissingFromNull(t *testing.T) {
	result, err := Compare(Input{
		Left:  NamedDictionary{Properties: Dictionary{"removed": Null()}},
		Right: NamedDictionary{Properties: Dictionary{"added": Null()}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Properties) != 2 {
		t.Fatalf("properties = %d, want 2", len(result.Properties))
	}
	if propertyByName(t, result, "added").Kind != ChangeAdded ||
		propertyByName(t, result, "removed").Kind != ChangeRemoved {
		t.Fatalf("missing/null changes = %#v", result.Properties)
	}
}

func TestCompareSplitsSeparatedChangesIntoContextualChunks(t *testing.T) {
	left := "zero\none\nleft first\nthree\nfour\nfive\nsix\nleft second\neight\nnine"
	right := "zero\none\nright first\nthree\nfour\nfive\nsix\nright second\neight\nnine"
	result, err := Compare(Input{
		ContextLines: 2,
		Left:         NamedDictionary{Properties: Dictionary{"value": String(left)}},
		Right:        NamedDictionary{Properties: Dictionary{"value": String(right)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	property := propertyByName(t, result, "value")
	if len(property.Chunks) != 2 {
		t.Fatalf("chunks = %d, want 2: %#v", len(property.Chunks), property.Chunks)
	}
	for index, chunk := range property.Chunks {
		changed := 0
		for _, row := range chunk.Rows {
			if row.Kind != LineEqual {
				changed++
			}
		}
		if changed != 1 {
			t.Fatalf("chunk %d has %d changed rows, want 1", index, changed)
		}
	}
}

func TestCompareKeepsInlineEqualAndChangedSegments(t *testing.T) {
	result, err := Compare(Input{
		Left: NamedDictionary{Properties: Dictionary{
			"value": String("prefix left middle same suffix"),
		}},
		Right: NamedDictionary{Properties: Dictionary{
			"value": String("prefix right middle same suffix"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	row := propertyByName(t, result, "value").Chunks[0].Rows[0]
	if len(row.Left.Segments) < 3 || len(row.Right.Segments) < 3 {
		t.Fatalf("inline segments do not preserve equal context: left=%#v right=%#v", row.Left.Segments, row.Right.Segments)
	}
	if row.Left.Segments[0].Kind != LineEqual ||
		row.Left.Segments[1].Kind != LineRemoved ||
		row.Right.Segments[1].Kind != LineAdded {
		t.Fatalf("inline segment kinds = left:%#v right:%#v", row.Left.Segments, row.Right.Segments)
	}
}

func TestCompareTreatsEnumBooleanAndNumberValuesAtomically(t *testing.T) {
	result, err := Compare(Input{
		Left: NamedDictionary{Properties: Dictionary{
			"boolean": Boolean(true),
			"enum":    Enum("FIRST"),
			"number":  Number(json.Number("12")),
			"string":  String("true"),
		}},
		Right: NamedDictionary{Properties: Dictionary{
			"boolean": Boolean(false),
			"enum":    Enum("SECOND"),
			"number":  Number(json.Number("13")),
			"string":  String("false"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"boolean", "enum", "number"} {
		row := propertyByName(t, result, name).Chunks[0].Rows[0]
		for side, line := range map[string]*Line{"left": row.Left, "right": row.Right} {
			if line == nil || len(line.Segments) != 1 || line.Segments[0].Kind != LineChanged {
				t.Fatalf("%s %s value was not compared atomically: %#v", name, side, line)
			}
			if line.Segments[0].Text != line.Text {
				t.Fatalf("%s %s atomic segment = %q, want whole value %q", name, side, line.Segments[0].Text, line.Text)
			}
		}
	}
	stringRow := propertyByName(t, result, "string").Chunks[0].Rows[0]
	if len(stringRow.Left.Segments) == 1 && stringRow.Left.Segments[0].Kind == LineChanged {
		t.Fatalf("regular string was compared atomically: %#v", stringRow)
	}
}

func TestCompareJSONHintPrettifiesValidJSONAndFallsBackToRawString(t *testing.T) {
	result, err := Compare(Input{
		Left: NamedDictionary{Properties: Dictionary{
			"valid": {
				Type:      ValueString,
				CompareAs: CompareJSON,
				Raw:       `{"enabled":true}`,
			},
			"invalid": JSON(`{"broken"`),
		}},
		Right: NamedDictionary{Properties: Dictionary{
			"valid": {
				Type:      ValueString,
				CompareAs: CompareJSON,
				Raw:       `{"enabled":false}`,
			},
			"invalid": JSON(`{"different"`),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	valid := propertyByName(t, result, "valid")
	if valid.Left.CompareAs != CompareJSON ||
		!strings.Contains(valid.Left.Text, "\n  \"enabled\": true") ||
		!strings.Contains(valid.Right.Text, "\n  \"enabled\": false") {
		t.Fatalf("JSON comparison hint did not pretty-print valid input: %#v", valid)
	}
	invalid := propertyByName(t, result, "invalid")
	if invalid.Left.Text != `{"broken"` || invalid.Right.Text != `{"different"` {
		t.Fatalf("invalid JSON did not fall back to raw strings: %#v", invalid)
	}
}

func TestCompareRejectsInvalidNumber(t *testing.T) {
	_, err := Compare(Input{
		Left:  NamedDictionary{Properties: Dictionary{"count": Number(json.Number("NaN"))}},
		Right: NamedDictionary{},
	})
	if err == nil || !strings.Contains(err.Error(), `left property "count" has invalid number`) {
		t.Fatalf("error = %v, want contextual invalid number error", err)
	}
}

func propertyByName(t *testing.T, result Result, name string) Property {
	t.Helper()
	for _, property := range result.Properties {
		if property.Name == name {
			return property
		}
	}
	t.Fatalf("property %q not found in %#v", name, result.Properties)
	return Property{}
}
