package viewutil

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestCollapsedSidebarTitleFallbacks(t *testing.T) {
	tests := []struct {
		height int
		title  string
	}{
		{4, ""},
		{5, "C"},
		{6, "Cx"},
		{7, "Ctx"},
		{10, "Ctx"},
		{11, "Context"},
	}

	for _, tt := range tests {
		view := ansi.Strip(CollapsedSidebarPanel(tt.height, true, true, "Context", "Ctx", "Cx", "C"))
		lines := strings.Split(view, "\n")
		if len(lines) != tt.height {
			t.Fatalf("height %d rendered %d rows: %q", tt.height, len(lines), view)
		}
		var title strings.Builder
		for _, line := range lines {
			if letter, ok := strings.CutPrefix(line, "▐"); ok {
				title.WriteString(strings.TrimSuffix(letter, "▌"))
			}
		}
		if title.String() != tt.title {
			t.Errorf("height %d title = %q, want %q", tt.height, title.String(), tt.title)
		}
		if tt.title == "" && view != "─╮\n │ \n │ \n─╯" {
			t.Errorf("short rail = %q, want a continuous border with no title", view)
		}
	}
}

func TestCollapsedSidebarTruncatesTitleWithoutAlternatives(t *testing.T) {
	view := ansi.Strip(CollapsedSidebarPanel(8, false, false, "Projects"))
	if !strings.Contains(view, " P \n r \n o \n ⋮ ") {
		t.Fatalf("collapsed Projects title = %q, want Pro and a vertical ellipsis", view)
	}
}
