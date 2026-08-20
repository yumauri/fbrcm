package harness

import "testing"

func TestModes(t *testing.T) {
	tests := []struct {
		name             string
		mode             Mode
		cassetteExists   bool
		snapshotExists   bool
		wantCapture      bool
		wantUpdateOutput bool
	}{
		{"replay", ModeReplay, true, true, false, false},
		{"record missing", ModeRecordMissing, false, false, true, true},
		{"record missing preserves existing", ModeRecordMissing, true, true, false, false},
		{"refresh http", ModeRefreshHTTP, true, true, true, false},
		{"update output", ModeUpdateOutput, true, true, false, true},
		{"refresh all", ModeRefreshAll, true, true, true, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.mode.Capture(test.cassetteExists); got != test.wantCapture {
				t.Fatalf("Capture() = %v, want %v", got, test.wantCapture)
			}
			if got := test.mode.UpdateOutput(test.snapshotExists); got != test.wantUpdateOutput {
				t.Fatalf("UpdateOutput() = %v, want %v", got, test.wantUpdateOutput)
			}
		})
	}
}

func TestParseModeRejectsUnknownMode(t *testing.T) {
	if _, err := ParseMode("unknown"); err == nil {
		t.Fatal("ParseMode() accepted an unknown mode")
	}
}
