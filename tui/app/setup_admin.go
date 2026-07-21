package app

import (
	"fmt"
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/config"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/components/setup"
)

type profileRenameCompletedMsg struct{ err error }

func (m *Model) openAuthDeleteDialog(request setup.AuthDeleteRequestedMsg) {
	body := []string{
		"Authentication: " + dialogParameterNameStyle.Render(request.AuthID),
		"",
		"This removes the identity and its stored credential",
		"and token files.",
	}
	if request.BoundProjects > 0 {
		body = append(body,
			"",
			fmt.Sprintf("This identity is bound to %s.", rcdisplay.FormatCount(request.BoundProjects, "cached project", "cached projects")),
			"They will require rebinding before Firebase operations can continue.",
		)
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Delete Authentication?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Delete", Variant: dialogcmp.ButtonVariantDanger, OnPress: func() tea.Msg {
				return setup.AuthDeleteConfirmedMsg{AuthID: request.AuthID}
			}},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

func (m *Model) openProfileDeleteDialog(request setup.ProfileDeleteRequestedMsg) {
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Delete Profile?",
		Body: []string{
			"Profile: " + dialogParameterNameStyle.Render(request.Profile),
			"",
			"This permanently removes its authentication, projects, drafts, and caches.",
			request.ConfigPath,
			request.CachePath,
		},
		Buttons: []dialogcmp.Button{
			{Label: "Delete", Variant: dialogcmp.ButtonVariantDanger, OnPress: func() tea.Msg {
				return setup.ProfileDeleteConfirmedMsg{Profile: request.Profile}
			}},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

func (m *Model) openSetupErrorDialog(request setup.ErrorRequestedMsg) {
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: request.Title,
		Body:  request.Body,
		Buttons: []dialogcmp.Button{
			{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

func (m *Model) openProfileRenameInput(request setup.ProfileRenameRequestedMsg) tea.Cmd {
	anchor, ok := m.setup.ProfileRenamePosition(m.width, m.height)
	if !ok || anchor.Profile != request.Profile {
		return nil
	}
	m.closeOverlays()
	m.profileRename = &profileRenameSession{profile: request.Profile}
	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Open(anchor.X, anchor.Y, anchor.Width, anchor.MaxWidth, request.Profile)
	return cmd
}

func (m *Model) submitProfileRenameInput() tea.Cmd {
	session := m.profileRename
	if session == nil {
		return nil
	}
	name := m.renameInput.Value()
	if err := config.ValidateProfileName(name); err != nil {
		m.openSetupErrorDialog(setup.ErrorRequestedMsg{Title: "Rename Profile Failed", Body: []string{err.Error()}})
		return nil
	}
	profiles, err := config.ListProfiles()
	if err != nil {
		m.openSetupErrorDialog(setup.ErrorRequestedMsg{Title: "Rename Profile Failed", Body: []string{err.Error()}})
		return nil
	}
	if name != session.profile && slices.Contains(profiles, name) {
		m.openSetupErrorDialog(setup.ErrorRequestedMsg{
			Title: "Rename Profile Failed",
			Body:  []string{fmt.Sprintf("profile %q already exists", name)},
		})
		return nil
	}
	if name == session.profile {
		m.profileRename = nil
		m.closeRenameInput()
		return nil
	}
	oldName := session.profile
	m.profileRename = nil
	m.closeRenameInput()
	return func() tea.Msg {
		return profileRenameCompletedMsg{err: config.RenameProfile(oldName, name)}
	}
}

func (m Model) updateProfileRenameCompleted(msg profileRenameCompletedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.openSetupErrorDialog(setup.ErrorRequestedMsg{Title: "Rename Profile Failed", Body: []string{msg.err.Error()}})
		return m, nil
	}
	var cmd tea.Cmd
	m.setup, cmd = m.setup.OpenProfiles()
	return m, cmd
}
