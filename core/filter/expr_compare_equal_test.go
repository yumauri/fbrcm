package filter

import "testing"

func TestExprValuesCoerceEqual(t *testing.T) {
	tests := []struct {
		name  string
		left  any
		right any
		want  bool
	}{
		{name: "bool true string", left: "true", right: true, want: true},
		{name: "bool false string", left: "false", right: false, want: true},
		{name: "bool mismatch", left: "true", right: false, want: false},
		{name: "int number", left: "10", right: 10, want: true},
		{name: "float number", left: "3.5", right: 3.5, want: true},
		{name: "number mismatch", left: "10", right: 11, want: false},
		{name: "non string left", left: 10, right: "10", want: false},
		{name: "unsupported right type", left: "x", right: struct{}{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exprValuesCoerceEqual(tt.left, tt.right); got != tt.want {
				t.Fatalf("exprValuesCoerceEqual(%#v, %#v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestExprValuesEqualRootGroupAndAnyValue(t *testing.T) {
	root := rootGroup{}
	anyOn := anyValue{values: []any{"on"}}
	anyOff := anyValue{values: []any{"off"}}

	if !exprValuesEqual(root, nil) {
		t.Fatal("root group should equal nil")
	}
	if !exprValuesEqual(nil, root) {
		t.Fatal("nil should equal root group")
	}
	if !exprValuesEqual(root, rootGroupLabel) {
		t.Fatal("root group should equal root label string")
	}
	if !exprValuesEqual(anyOn, "on") {
		t.Fatal("anyValue should match contained literal")
	}
	if exprValuesEqual(anyOn, anyOff) {
		t.Fatal("distinct anyValue sets should not match")
	}
}

func TestMatchParameterRootGroupLabel(t *testing.T) {
	cfg := exprTestConfig()
	expr, err := CompileExpression(`group == "(root)"`)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	got, err := expr.MatchParameter("demo", "Demo", cfg, "feature_login", "")
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if !got {
		t.Fatal("root parameter should match group == (root)")
	}
}
