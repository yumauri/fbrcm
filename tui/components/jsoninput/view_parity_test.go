package jsoninput

import (
	"testing"

	"github.com/yumauri/fbrcm/tui/testutil"
)

func parityTestModel() Model {
	m, _ := New().Open(40, 20, `{"enabled":true,"count":3}`)
	return m
}

func TestJSONInputViewSnapshot(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := testutil.NormalizeViewSnapshot(parityTestModel().View())
	if got != jsonInputViewSnapshot {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, jsonInputViewSnapshot)
	}
}

const jsonInputViewSnapshot = `╭──────────────────────────────────╮
│                                  │
│  1 {                             │
│  2   "enabled": true,            │
│  3   "count": 3                  │
│  4 }                             │
│                                  │
│                                  │
│                                  │
│                                  │
│                                  │
│                                  │
│                                  │
│                                  │
│  ctrl+s/ctrl+enter save …        │
╰──────────────────────────────────╯`
