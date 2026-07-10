package draft

import (
	"encoding/json"
	"testing"
)

func TestMergeWithLatest(t *testing.T) {
	tests := []struct {
		name       string
		base       json.RawMessage
		draft      json.RawMessage
		latest     json.RawMessage
		wantChange bool
		wantValue  string
		wantErr    bool
		wantAbsent bool
	}{
		{
			name:       "keeps local change when remote unchanged",
			base:       remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:      remoteConfigRaw("1", map[string]string{"flag": "local"}),
			latest:     remoteConfigRaw("2", map[string]string{"flag": "old"}),
			wantChange: true,
			wantValue:  "local",
		},
		{
			name:       "obsolete draft removed when only remote changed",
			base:       remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:      remoteConfigRaw("1", map[string]string{"flag": "old"}),
			latest:     remoteConfigRaw("2", map[string]string{"flag": "remote"}),
			wantChange: false,
		},
		{
			name:    "reports conflict when local and remote both changed",
			base:    remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:   remoteConfigRaw("1", map[string]string{"flag": "local"}),
			latest:  remoteConfigRaw("2", map[string]string{"flag": "remote"}),
			wantErr: true,
		},
		{
			name:       "keeps local deletion when remote unchanged",
			base:       remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:      remoteConfigRaw("1", map[string]string{}),
			latest:     remoteConfigRaw("2", map[string]string{"flag": "old"}),
			wantChange: true,
			wantAbsent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := MergeWithLatest(tt.base, tt.draft, tt.latest)
			if tt.wantErr {
				if err == nil {
					t.Fatal("MergeWithLatest returned nil error")
				}
				return
			}
			if err != nil {
				t.Fatalf("MergeWithLatest returned error: %v", err)
			}
			if changed != tt.wantChange {
				t.Fatalf("changed = %v, want %v", changed, tt.wantChange)
			}
			if !changed {
				return
			}
			if tt.wantAbsent {
				assertFlagMissing(t, got)
				return
			}
			assertParamValue(t, got, "flag", tt.wantValue)
		})
	}
}
