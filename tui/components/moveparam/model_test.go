package moveparam

import (
	"testing"
	"time"
)

func TestMoveWrapsAcrossOptionsAndInput(t *testing.T) {
	m := New().Open(4, 5, "parameter", []Option{
		{Key: "alpha", Label: "Alpha"},
		{Key: "", Label: "(root)"},
	})

	if got := m.rowsCount(); got != 3 {
		t.Fatalf("rowsCount() = %d, want 3", got)
	}
	m.Move(-1)
	if got, ok := m.Current(); !ok || got.Key != "" {
		t.Fatalf("Move(-1) should wrap to root option: %#v, %v", got, ok)
	}
	m.Move(1)
	m.Move(1)
	if !m.InputSelected() {
		t.Fatal("second row should be input")
	}
	m.input.SetValue(" new-group ")
	if got, ok := m.Current(); !ok || got.Key != "new-group" || got.Label != "new-group" {
		t.Fatalf("Current() = %#v, %v", got, ok)
	}
	m.Move(1)
	if got, ok := m.Current(); !ok || got.Key != "" {
		t.Fatalf("next Current() = %#v, %v", got, ok)
	}
}

func TestTypeaheadAccumulatesThenResetsAfterTimeout(t *testing.T) {
	m := New().Open(0, 0, "parameter", []Option{
		{Key: "alpha", Label: "Alpha"},
		{Key: "alpine", Label: "Alpine"},
		{Key: "beta", Label: "Beta"},
	})
	now := time.Unix(100, 0)

	if !m.Typeahead("a", now) || m.selected != 0 {
		t.Fatalf("first typeahead selected %d, want 0", m.selected)
	}
	if !m.Typeahead("l", now.Add(time.Millisecond)) || m.search != "al" {
		t.Fatalf("accumulated search = %q, selected = %d", m.search, m.selected)
	}
	if !m.Typeahead("b", now.Add(typeaheadTimeout+time.Millisecond)) || m.selected != 2 || m.search != "b" {
		t.Fatalf("reset typeahead search = %q, selected = %d", m.search, m.selected)
	}
}

func TestEmptyOptionsProduceNoOverlay(t *testing.T) {
	m := New().Open(2, 3, "parameter", nil)
	if m.HeaderView() != "" || m.ListView() != "" {
		t.Fatal("empty options should render no overlay")
	}
	if _, ok := m.Current(); ok {
		t.Fatal("empty options should have no current option")
	}
}
