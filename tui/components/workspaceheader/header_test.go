package workspaceheader

import "testing"

func TestTabAtMatchesRenderedTabWidths(t *testing.T) {
	for _, test := range []struct {
		width    int
		selected int
	}{{80, 0}, {18, 1}, {18, 2}} {
		widths := tabWidths(test.width, test.selected, keys())
		x := min(2, test.width)
		for index, tabWidth := range widths {
			for column := x; column < x+tabWidth; column++ {
				if got, ok := TabAt(test.width, test.selected, column); !ok || got != index {
					t.Fatalf("TabAt(width=%d, selected=%d, x=%d) = (%d, %v), want (%d, true)", test.width, test.selected, column, got, ok, index)
				}
			}
			x += tabWidth
			if index < tabCount-1 {
				for column := x; column < x+2; column++ {
					if got, ok := TabAt(test.width, test.selected, column); ok {
						t.Fatalf("separator TabAt(width=%d, selected=%d, x=%d) = (%d, true), want no tab", test.width, test.selected, column, got)
					}
				}
				x += 2
			}
		}
	}
}
