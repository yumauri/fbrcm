package app

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/contextpanel"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

const contextPauseMark = " ⏸"

func (m Model) contextSidebarHeight(layout panelLayout) int {
	if m.promote.WorkspaceOpen() {
		return 0
	}
	const borderRows, inputRows = 2, 2
	maxHeight := layout.topHeight * 2 / 3
	maxInputs := max((maxHeight-borderRows)/inputRows, 0)
	inputs := min(max(m.requestContext.list.Len(), 1), maxInputs)
	return min(borderRows+inputRows*inputs, maxHeight)
}

func (m Model) contextLayout(popup bool) contextpanel.Layout {
	c := m.requestContext
	title := "Context"
	if c.paused {
		title += contextPauseMark
	}
	if !popup {
		layout := m.workspaceLayout()
		height := m.contextSidebarHeight(layout)
		return contextpanel.Layout{
			Y: layout.topHeight - height, Width: layout.leftWidth, Height: height,
			Title: title, Empty: "No inputs", Sidebar: true, Count: c.list.Len(),
			TitleKey: tuiconfig.ActionKeyHint(tuiconfig.BlockContext, tuiconfig.ActionFocusContext),
		}
	}

	width := min(m.width-4, 76)
	height := min(3+2*max(m.contextList().Len(), 1), max(m.height-4, 0))
	titleKey := ""
	count := 0
	bindings := []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockContext, tuiconfig.ActionSubmit, "choose"),
		tuiconfig.Binding(tuiconfig.BlockContext, tuiconfig.ActionCancel, "cancel"),
	}
	switch c.overlay {
	case contextFields:
		title = "Add Context Input"
	case contextNames:
		title = contextNameTitle(c.nameField)
	case contextRequests:
		title = "Choose a Firebase Fetch"
	default:
		titleKey = tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionOpenContext)
		count = c.list.Len()
		pauseLabel := "pause"
		if c.paused {
			pauseLabel = "resume"
		}
		bindings = append(contextHelp(pauseLabel), tuiconfig.Binding(tuiconfig.BlockContext, tuiconfig.ActionCancel, "close"))
	}
	return contextpanel.Layout{
		X: (m.width - width) / 2, Y: (m.height - height) / 2, Width: width, Height: height,
		Title: title, TitleKey: titleKey, Count: count, Empty: "No inputs",
		Footer: " " + viewutil.ShortHelpView(max(width-4, 0), bindings...),
	}
}

func (m Model) contextBadge() (string, int) {
	if m.promote.WorkspaceOpen() || (m.requestContext.scenario == nil && !m.requestContext.paused && m.projectsMode != projectsPanelModeCollapsed) {
		return "", 0
	}
	label := "Context"
	style := styles.TitleStyle(true)
	if m.requestContext.paused {
		label += contextPauseMark
		style = styles.PanelTitleInactiveTab
	} else if m.requestContext.scenario != nil {
		label += " on"
	}
	badge := style.Padding(0, 1).Render(label)
	x := m.width - lipgloss.Width(profileBadgeView(m.profileName, m.width)) - lipgloss.Width(badge) - 1
	return badge, max(x, 0)
}

func (m Model) contextBadgeAt(x, y int) bool {
	if y != 0 || m.detailsVisible {
		return false
	}
	badge, left := m.contextBadge()
	return badge != "" && x >= left && x < left+lipgloss.Width(badge)
}

func (m Model) appendContextLayers(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if badge, x := m.contextBadge(); badge != "" && !m.detailsVisible {
		layers = append(layers, lipgloss.NewLayer(badge).ID("request-context-badge").X(x).Y(0).Z(1))
	}
	if m.requestContext.overlay != contextInline {
		layout := m.contextLayout(true)
		layers = append(layers, lipgloss.NewLayer(m.contextList().View(layout, true)).
			ID("request-context").X(layout.X).Y(layout.Y).Z(2))
	}
	return layers
}
