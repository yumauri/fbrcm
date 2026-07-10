package value

import (
	"strings"
	"testing"
)

func TestIsJSONNumber(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "integer", in: "42", want: true},
		{name: "negative decimal", in: "-3.14", want: true},
		{name: "exponent", in: "1e-9", want: true},
		{name: "trimmed", in: " 0.5 ", want: true},
		{name: "empty", in: "", want: false},
		{name: "leading zero", in: "01", want: false},
		{name: "nan", in: "NaN", want: false},
		{name: "trailing text", in: "1px", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJSONNumber(tt.in); got != tt.want {
				t.Fatalf("IsJSONNumber(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidRawValueForType(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		valueType string
		want      bool
	}{
		{name: "string empty type", value: "hello", valueType: "", want: true},
		{name: "boolean true", value: "true", valueType: "BOOLEAN", want: true},
		{name: "boolean false", value: "false", valueType: "BOOLEAN", want: true},
		{name: "boolean uppercase rejected", value: "True", valueType: "BOOLEAN", want: false},
		{name: "number integer", value: "42", valueType: "NUMBER", want: true},
		{name: "number leading zero rejected", value: "01", valueType: "NUMBER", want: false},
		{name: "number nan rejected", value: "NaN", valueType: "NUMBER", want: false},
		{name: "json object", value: `{"a":1}`, valueType: "JSON", want: true},
		{name: "json invalid", value: `{`, valueType: "JSON", want: false},
		{name: "unknown type", value: "x", valueType: "UNKNOWN", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidRawValueForType(tt.value, tt.valueType); got != tt.want {
				t.Fatalf("ValidRawValueForType(%q, %q) = %v, want %v", tt.value, tt.valueType, got, tt.want)
			}
		})
	}
}

func TestValidateRawValueForType(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		valueType string
		wantErr   string
	}{
		{name: "string ok", value: "hello", valueType: "STRING"},
		{name: "boolean lowercase", value: "true", valueType: "BOOLEAN"},
		{name: "boolean uppercase", value: "True", valueType: "BOOLEAN"},
		{name: "boolean invalid", value: "yes", valueType: "BOOLEAN", wantErr: "invalid boolean"},
		{name: "number ok", value: "42", valueType: "NUMBER"},
		{name: "number leading zero rejected", value: "01", valueType: "NUMBER", wantErr: "invalid number"},
		{name: "json ok", value: "{}", valueType: "JSON"},
		{name: "json invalid", value: "{", valueType: "JSON", wantErr: "invalid json"},
		{name: "unknown type", value: "x", valueType: "UNKNOWN", wantErr: "invalid value type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRawValueForType(tt.value, tt.valueType)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateRawValueForType() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateRawValueForType() error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidateRawValueForType() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}
