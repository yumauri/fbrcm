package target

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		input     string
		kind      Kind
		projectID string
		canonical string
	}{
		{input: "demo", kind: Client, projectID: "demo", canonical: "demo"},
		{input: "client@demo", kind: Client, projectID: "demo", canonical: "demo"},
		{input: "CLIENT@demo", kind: Client, projectID: "demo", canonical: "demo"},
		{input: "server@demo", kind: Server, projectID: "demo", canonical: "server@demo"},
		{input: "SERVER@ demo ", kind: Server, projectID: "demo", canonical: "server@demo"},
		{input: "name@example", kind: Client, projectID: "name@example", canonical: "name@example"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if got.Kind != tt.kind || got.ProjectID != tt.projectID || got.String() != tt.canonical {
				t.Fatalf("Parse(%q) = %#v / %q", tt.input, got, got.String())
			}
		})
	}
}

func TestParseRejectsMissingProject(t *testing.T) {
	for _, input := range []string{"", "client@", "server@  "} {
		if _, err := Parse(input); err == nil {
			t.Fatalf("Parse(%q) succeeded", input)
		}
	}
}

func TestParseSelectorReportsExplicitKind(t *testing.T) {
	for input, want := range map[string]bool{
		"demo":        false,
		"client@demo": true,
		"server@demo": true,
	} {
		_, got, err := ParseSelector(input)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("ParseSelector(%q) explicit = %t, want %t", input, got, want)
		}
	}
}

func TestExactFilter(t *testing.T) {
	for input, want := range map[string]string{
		"demo":        "=demo",
		"client@demo": "=demo",
		"server@demo": "server@=demo",
	} {
		got, err := ExactFilter(input)
		if err != nil || got != want {
			t.Fatalf("ExactFilter(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}
