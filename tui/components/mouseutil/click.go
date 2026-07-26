package mouseutil

import "time"

const DoubleClickWindow = 400 * time.Millisecond

// ClickTracker reports whether successive clicks target the same selectable
// item within the application's double-click window.
type ClickTracker struct {
	kind  int
	index int
	at    time.Time
	set   bool
}

func (t *ClickTracker) Register(kind, index int, at time.Time) bool {
	double := t.set &&
		t.kind == kind &&
		t.index == index &&
		!at.Before(t.at) &&
		at.Sub(t.at) <= DoubleClickWindow
	t.kind, t.index, t.at, t.set = kind, index, at, true
	return double
}

func (t *ClickTracker) Reset() {
	*t = ClickTracker{}
}
