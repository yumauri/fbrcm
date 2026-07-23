package viewutil

import "testing"

func TestScrollbarState(t *testing.T) {
	if got := ScrollbarState(5, 0, 5); got.Visible {
		t.Fatalf("fitting content has visible scrollbar: %#v", got)
	}
	top := ScrollbarState(20, 0, 5)
	bottom := ScrollbarState(20, 15, 5)
	if !top.Visible || top.ThumbStart != 0 {
		t.Fatalf("top scrollbar = %#v", top)
	}
	if !bottom.Visible || bottom.ThumbEnd != 4 {
		t.Fatalf("bottom scrollbar = %#v", bottom)
	}
}
