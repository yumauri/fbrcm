package shared

import (
	"strings"
	"testing"
)

func TestReadRemoteConfigInput(t *testing.T) {
	cfg, raw, err := ReadRemoteConfigInput(strings.NewReader(`{"parameters":{"flag":{"defaultValue":{"value":"on"}}}}`))
	if err != nil {
		t.Fatalf("ReadRemoteConfigInput returned error: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("remote config raw is empty")
	}
	if cfg.Parameters["flag"].DefaultValue.Value != "on" {
		t.Fatalf("flag value = %q, want on", cfg.Parameters["flag"].DefaultValue.Value)
	}
}

func TestReadRemoteConfigInputWrappedCachePayload(t *testing.T) {
	cfg, raw, err := ReadRemoteConfigInput(strings.NewReader(`{"remote_config":{"parameters":{"flag":{"defaultValue":{"value":"on"}}}}}`))
	if err != nil {
		t.Fatalf("ReadRemoteConfigInput returned error: %v", err)
	}
	if !strings.Contains(string(raw), `"parameters"`) {
		t.Fatalf("remote config raw = %s, want extracted remote_config", raw)
	}
	if cfg.Parameters["flag"].DefaultValue.Value != "on" {
		t.Fatalf("flag value = %q, want on", cfg.Parameters["flag"].DefaultValue.Value)
	}
}

func TestReadRemoteConfigInputRejectsInvalidJSON(t *testing.T) {
	_, _, err := ReadRemoteConfigInput(strings.NewReader(`{`))
	if err == nil {
		t.Fatalf("ReadRemoteConfigInput accepted invalid JSON")
	}
	if got, want := err.Error(), "stdin remote config is not valid json"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}
