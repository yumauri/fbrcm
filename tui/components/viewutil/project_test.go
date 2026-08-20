package viewutil

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestProjectLineUsesNameAndParenthesizedID(t *testing.T) {
	project := core.Project{Name: "Example Mobile", ProjectID: "example-mobile-a1b2"}
	if got, want := ansi.Strip(ProjectLine(project)), "Project: Example Mobile (example-mobile-a1b2)"; got != want {
		t.Fatalf("ProjectLine = %q, want %q", got, want)
	}
	if got := ProjectLine(project); !strings.Contains(got, styles.TreeProjectName.Render(project.Name)) {
		t.Fatalf("ProjectLine does not use the shared colored project-name style: %q", got)
	}
}
