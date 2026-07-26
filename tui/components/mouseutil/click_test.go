package mouseutil

import (
	"testing"
	"time"
)

func TestClickTrackerRequiresSameTargetWithinWindow(t *testing.T) {
	now := time.Unix(100, 0)
	var tracker ClickTracker
	if tracker.Register(1, 2, now) {
		t.Fatal("first click reported a double click")
	}
	if tracker.Register(1, 3, now.Add(time.Millisecond)) {
		t.Fatal("different item reported a double click")
	}
	if !tracker.Register(1, 3, now.Add(2*time.Millisecond)) {
		t.Fatal("same item within window did not report a double click")
	}
	if tracker.Register(1, 3, now.Add(DoubleClickWindow+3*time.Millisecond)) {
		t.Fatal("click outside the window reported a double click")
	}
}
