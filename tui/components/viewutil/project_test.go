package viewutil

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestProjectLineUsesNameAndParenthesizedID(t *testing.T) {
	project := core.Project{Name: "Mercato Mobile", ProjectID: "mercato-mobile-9eac5"}
	if got, want := ansi.Strip(ProjectLine(project)), "Project: Mercato Mobile (mercato-mobile-9eac5)"; got != want {
		t.Fatalf("ProjectLine = %q, want %q", got, want)
	}
	if got := ProjectLine(project); !strings.Contains(got, styles.TreeProjectName.Render(project.Name)) {
		t.Fatalf("ProjectLine does not use the shared colored project-name style: %q", got)
	}
}
