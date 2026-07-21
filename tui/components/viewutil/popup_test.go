package viewutil

import "testing"

func TestPopupContentLineUsesSharedInsets(t *testing.T) {
	if got, want := PopupContentLine("body", 6), "  body   "; got != want {
		t.Fatalf("PopupContentLine = %q, want %q", got, want)
	}
	if got, want := PopupInnerWidth(6), 9; got != want {
		t.Fatalf("PopupInnerWidth = %d, want %d", got, want)
	}
}
