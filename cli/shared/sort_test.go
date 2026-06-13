package shared

import "testing"

func TestSortedStringKeys(t *testing.T) {
	got := SortedStringKeys(map[string]int{
		"beta":  2,
		"alpha": 1,
		"Gamma": 3,
	})
	want := []string{"Gamma", "alpha", "beta"}

	if len(got) != len(want) {
		t.Fatalf("SortedStringKeys length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SortedStringKeys = %#v, want %#v", got, want)
		}
	}
}
