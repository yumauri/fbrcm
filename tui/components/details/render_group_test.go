package details

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

func TestGroupDetailsRendersAndStagesMetadata(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetGroupData(&messages.GroupViewData{
		Project:    core.Project{Name: "Demo", ProjectID: "demo"},
		Group:      core.ParametersGroup{Key: "checkout", Label: "checkout", Description: "Checkout flags", Parameters: []core.ParametersEntry{{Key: "enabled"}}},
		GroupNames: []string{"checkout", "other"},
	})
	view := ansi.Strip(m.View())
	for _, want := range []string{"Details", "Name", "checkout", "Description", "Checkout flags", "Parameters", "1"} {
		if !strings.Contains(view, want) {
			t.Fatalf("group Details missing %q:\n%s", want, view)
		}
	}
	m.nameInput.SetValue("payments")
	m.descInput.SetValue("Payment flags")
	edit, ok := m.GroupEdit()
	if !ok || edit.Create || edit.Name != "checkout" || edit.NextName != "payments" || edit.NextDescription != "Payment flags" {
		t.Fatalf("GroupEdit = %#v, %v", edit, ok)
	}
}

func TestNewGroupDetailsStagesCreation(t *testing.T) {
	m := New().SetGroupData(&messages.GroupViewData{
		Project:    core.Project{Name: "Demo", ProjectID: "demo"},
		GroupNames: []string{"existing"},
	})
	m.nameInput.SetValue("new-group")
	m.descInput.SetValue("Metadata only")

	edit, ok := m.GroupEdit()
	if !ok || !edit.Create || edit.Name != "" || edit.NextName != "new-group" || edit.NextDescription != "Metadata only" {
		t.Fatalf("GroupEdit = %#v, %v", edit, ok)
	}
}
