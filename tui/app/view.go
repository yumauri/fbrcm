package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/components/minsize"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	rootStyle = lipgloss.NewStyle()

	offlineBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("196")).
				Padding(0, 1)

	profileBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.PaletteSlateBright).
				Background(styles.PaletteBlueDeep).
				Padding(0, 1)
)

func (m Model) View() tea.View {
	if m.width < minsize.MinWidth || m.height < minsize.MinHeight {
		body := m.profileOverlay(minsize.View(m.width, m.height))
		return appView(rootStyle.Render(body), tea.MouseModeNone)
	}
	if m.setup.IsOpen() && !m.setup.IsPopup() && !m.oauthDialog.IsOpen() {
		body := m.profileOverlay(m.setup.View(m.width, m.height))
		return appView(rootStyle.Render(body), tea.MouseModeAllMotion)
	}

	body := m.baseView()
	if m.setup.IsOpen() && !m.setup.IsPopup() {
		body = m.setup.View(m.width, m.height)
	}
	layers := m.overlayLayers(body)
	if len(layers) > 1 {
		body = lipgloss.NewCompositor(layers...).Render()
	}

	return appView(rootStyle.Render(body), m.mouseMode())
}

func appView(content string, mouseMode tea.MouseMode) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = mouseMode
	return v
}

func (m Model) mouseMode() tea.MouseMode {
	if m.workspaceMenu {
		return tea.MouseModeAllMotion
	}
	if m.helpPalette.IsOpen() {
		return tea.MouseModeAllMotion
	}
	if m.setup.IsOpen() {
		return tea.MouseModeAllMotion
	}
	if m.dialog.IsOpen() || m.oauthDialog.IsOpen() || m.authPicker.IsOpen() || m.projectIO.IsOpen() {
		return tea.MouseModeAllMotion
	}
	if m.active == panels.Logs {
		return tea.MouseModeNone
	}
	return tea.MouseModeAllMotion
}

func (m Model) baseView() string {
	popupOpen := m.popupWindowOpen()
	projectsActive := m.active == panels.Projects && !popupOpen
	logsActive := m.active == panels.Logs && !popupOpen
	rightPanel := m.workspacePanelView(popupOpen)
	if m.promote.WorkspaceOpen() {
		rightPanel = m.promote.ViewWithBorder(m.active == panels.Promote && !popupOpen, m.active == panels.Promote && !popupOpen)
	}
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.projects.ViewWithBorder(projectsActive, projectsActive && !popupOpen),
		rightPanel,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topRow,
		m.logs.ViewWithBorder(logsActive, logsActive && !popupOpen),
		m.helpView(),
	)
}

func (m Model) workspacePanelView(popupOpen bool) string {
	selected := m.selectedParametersTab()
	active := m.active == selected && !popupOpen
	previewTab, previewOpen := 0, false
	if header, _, ok := m.workspaceBaseHeaderLayout(); ok {
		previewTab, previewOpen = m.workspaceMenuPreview(header)
	}

	switch selected {
	case panels.Conditions:
		if previewOpen {
			return m.conditions.ViewWithWorkspacePreview(active, active, previewTab)
		}
		return m.conditions.ViewWithBorder(active, active)
	case panels.ABTests:
		if previewOpen {
			return m.abTests.ViewWithWorkspacePreview(active, active, previewTab)
		}
		return m.abTests.ViewWithBorder(active, active)
	case panels.Personalizations:
		if previewOpen {
			return m.personalizations.ViewWithWorkspacePreview(active, active, previewTab)
		}
		return m.personalizations.ViewWithBorder(active, active)
	case panels.Rollouts:
		if previewOpen {
			return m.rollouts.ViewWithWorkspacePreview(active, active, previewTab)
		}
		return m.rollouts.ViewWithBorder(active, active)
	default:
		if previewOpen {
			return m.parameters.ViewWithWorkspacePreview(active, active, previewTab)
		}
		return m.parameters.ViewWithBorder(active, active)
	}
}

func (m Model) popupWindowOpen() bool {
	return m.helpPalette.IsOpen() || m.contextOverlayOpen()
}

func (m Model) contextOverlayOpen() bool {
	return (m.setup.IsOpen() && m.setup.IsPopup()) ||
		m.parameters.HistoryPickerOpen() ||
		m.details.DropdownOpen() ||
		m.dialog.IsOpen() ||
		m.oauthDialog.IsOpen() ||
		m.diffView.IsOpen() ||
		m.boolPicker.IsOpen() ||
		m.jsonInput.IsOpen() ||
		m.numberInput.IsOpen() ||
		m.stringInput.IsOpen() ||
		m.moveParam.IsOpen() ||
		m.authPicker.IsOpen() ||
		m.renameInput.IsOpen() ||
		m.projectIO.IsOpen()
}

func (m Model) overlayLayers(body string) []*lipgloss.Layer {
	layers := []*lipgloss.Layer{lipgloss.NewLayer(body).ID("base")}
	layers = m.appendPromotePickerLayers(layers)
	layers = m.appendDetailsLayers(layers)
	layers = m.appendWorkspaceMenuLayers(layers)
	layers = m.appendHistoryPickerLayer(layers)
	layers = m.appendInputLayers(layers)
	layers = m.appendDiffViewLayer(layers)
	layers = m.appendDialogLayers(layers)
	layers = m.appendProjectIOLayer(layers)
	layers = m.appendSetupLayer(layers)
	layers = m.appendOAuthDialogLayer(layers)
	layers = m.appendOfflineLayer(layers)
	layers = m.appendHelpPaletteLayer(layers)
	layers = m.appendProfileLayer(layers)
	return layers
}

func (m Model) appendWorkspaceMenuLayers(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.workspaceMenu {
		return layers
	}
	header, panelX, ok := m.workspaceHeaderLayout()
	if !ok {
		return layers
	}
	geometry, ok := header.PopupGeometry()
	if !ok {
		return layers
	}
	active := m.workspaceMenuBorderActive()
	border := styles.BorderStyle(active)
	const z = 4
	layers = append(layers, lipgloss.NewLayer(header.PopupViewFocused(m.workspaceCursor, active, border)).
		ID("workspace-menu").
		X(panelX+geometry.X).
		Y(0).
		Z(z))
	return layers
}

func (m Model) workspaceMenuBorderActive() bool {
	return isWorkspacePanel(m.active)
}

func (m Model) appendDiffViewLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.diffView.IsOpen() {
		return layers
	}
	x, y := m.diffView.Position()
	return append(layers, lipgloss.NewLayer(m.diffView.View()).ID("diff-view").X(x).Y(y).Z(8))
}

func (m Model) appendPromotePickerLayers(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.promote.TargetPickerOpen() {
		return layers
	}
	sourceX, sourceY := m.promote.SourcePosition()
	layers = append(layers, lipgloss.NewLayer(m.promote.SourceView()).ID("promote-source").X(sourceX).Y(sourceY).Z(2))
	x, y := m.promote.TargetPosition()
	return append(layers, lipgloss.NewLayer(m.promote.TargetView()).ID("promote-target").X(x).Y(y).Z(3))
}

func (m Model) appendProjectIOLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.projectIO.IsOpen() {
		return layers
	}
	view := m.projectIO.View()
	x, y := m.projectIO.Position()
	layers = append(layers, lipgloss.NewLayer(view).ID("project-io").X(x).Y(y).Z(5))
	if m.projectIO.OptionSelectorOpen() {
		listX, listY := m.projectIO.OptionSelectorListPosition()
		layers = append(layers, lipgloss.NewLayer(m.projectIO.OptionSelectorListView()).ID("project-io-option-list").X(listX).Y(listY).Z(6))
		headerX, headerY := m.projectIO.OptionSelectorPosition()
		layers = append(layers, lipgloss.NewLayer(m.projectIO.OptionSelectorHeaderView()).ID("project-io-option-header").X(headerX).Y(headerY).Z(7))
	}
	return layers
}

func (m Model) appendSetupLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.setup.IsOpen() || !m.setup.IsPopup() {
		return layers
	}
	focused := !m.helpPalette.IsOpen() && !m.dialog.IsOpen() && !m.oauthDialog.IsOpen() && !m.renameInput.IsOpen()
	view := m.setup.PopupViewWithFocus(m.width, m.height, focused)
	return append(layers, lipgloss.NewLayer(view).
		ID("accounts-profiles").
		X(max((m.width-lipgloss.Width(view))/2, 0)).
		Y(max((m.height-lipgloss.Height(view))/2, 0)).
		Z(50))
}

func (m Model) appendHelpPaletteLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.helpPalette.IsOpen() {
		return layers
	}
	view := m.helpPaletteView()
	return append(layers, lipgloss.NewLayer(view).
		ID("help-palette").
		X(max((m.width-lipgloss.Width(view))/2, 0)).
		Y(max((m.height-lipgloss.Height(view))/2, 0)).
		Z(100))
}

func (m Model) appendHistoryPickerLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.parameters.HistoryPickerOpen() {
		return layers
	}
	x, y := m.parameters.HistoryPickerPosition()
	return append(layers, lipgloss.NewLayer(m.parameters.HistoryPickerView()).ID("history-version-picker").X(x).Y(y).Z(4))
}

func (m Model) appendDetailsLayers(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if m.detailsVisible {
		layers = append(layers, lipgloss.NewLayer(m.detailsPanelView()).ID("details").X(m.detailsX()).Y(0).Z(1))
		if m.details.DropdownOpen() {
			x, y := m.details.DropdownCurrentPosition()
			layers = append(layers, lipgloss.NewLayer(m.details.DropdownCurrentView()).ID("details-dropdown-current").X(x).Y(y).Z(2))
			x, y = m.details.DropdownListPosition()
			layers = append(layers, lipgloss.NewLayer(m.details.DropdownListView()).ID("details-dropdown-list").X(x).Y(y).Z(2))
		}
	}
	return layers
}

func (m Model) detailsPanelView() string {
	return m.details.ViewWithBorder(m.active == panels.Details && !m.popupWindowOpen())
}

func (m Model) appendInputLayers(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if m.boolPicker.IsOpen() {
		x, y := m.boolPicker.Position()
		layers = append(layers, lipgloss.NewLayer(m.boolPicker.View()).ID("bool-picker").X(x).Y(y).Z(2))
	}
	if m.jsonInput.IsOpen() {
		x, y := m.jsonInput.Position()
		layers = append(layers, lipgloss.NewLayer(m.jsonInput.View()).ID("json-input").X(x).Y(y).Z(3))
	}
	if m.numberInput.IsOpen() {
		x, y := m.numberInput.Position()
		layers = append(layers, lipgloss.NewLayer(m.numberInput.View()).ID("number-input").X(x).Y(y).Z(3))
	}
	if m.stringInput.IsOpen() {
		x, y := m.stringInput.Position()
		layers = append(layers, lipgloss.NewLayer(m.stringInput.View()).ID("string-input").X(x).Y(y).Z(3))
	}
	if m.moveParam.IsOpen() {
		listX, listY := m.moveParam.ListPosition()
		layers = append(layers, lipgloss.NewLayer(m.moveParam.ListView()).ID("move-list").X(listX).Y(listY).Z(2))
		x, y := m.moveParam.Position()
		layers = append(layers, lipgloss.NewLayer(m.moveParam.HeaderView()).ID("move-header").X(x).Y(y).Z(3))
	}
	if m.authPicker.IsOpen() {
		x, y := m.authPicker.Position()
		layers = append(layers, lipgloss.NewLayer(m.authPicker.View()).ID("auth-picker").X(x).Y(y).Z(4))
	}
	if m.renameInput.IsOpen() {
		x, y := m.renameInput.Position()
		z := 3
		if m.setup.IsOpen() && m.setup.IsPopup() {
			z = 60
		}
		layers = append(layers, lipgloss.NewLayer(m.renameInput.View()).ID("rename").X(x).Y(y).Z(z))
	}
	return layers
}

func (m Model) appendDialogLayers(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if m.dialog.IsOpen() {
		dialog := m.dialog.View()
		x, y := m.dialog.Position()
		z := 4
		if m.setup.IsOpen() && m.setup.IsPopup() {
			z = 60
		}
		layers = append(layers, lipgloss.NewLayer(dialog).ID("dialog").X(x).Y(y).Z(z))
	}
	return layers
}

func (m Model) appendOAuthDialogLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if !m.oauthDialog.IsOpen() {
		return layers
	}
	dialog := m.oauthDialog.View()
	x, y := m.oauthDialog.Position()
	return append(layers, lipgloss.NewLayer(dialog).ID("oauth-authorization").X(x).Y(y).Z(90))
}

func (m Model) appendOfflineLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	if firebase.IsOffline() {
		badge := offlineBadgeView()
		layers = append(layers, lipgloss.NewLayer(badge).ID("offline").X(max(m.width-lipgloss.Width(badge), 0)).Y(max(m.height-1, 0)).Z(99))
	}
	return layers
}

func (m Model) appendProfileLayer(layers []*lipgloss.Layer) []*lipgloss.Layer {
	badge := profileBadgeView(m.profileName, m.width)
	if badge == "" {
		return layers
	}
	return append(layers, lipgloss.NewLayer(badge).
		ID("profile").
		X(max(m.width-lipgloss.Width(badge), 0)).
		Y(0).
		Z(101))
}

func (m Model) profileBadgeAt(x, y int) bool {
	if y != 0 {
		return false
	}
	badge := profileBadgeView(m.profileName, m.width)
	if badge == "" {
		return false
	}
	left := max(m.width-lipgloss.Width(badge), 0)
	return x >= left && x < left+lipgloss.Width(badge)
}

func (m Model) profileOverlay(body string) string {
	layers := m.appendProfileLayer([]*lipgloss.Layer{lipgloss.NewLayer(body).ID("base")})
	if len(layers) == 1 {
		return body
	}
	return lipgloss.NewCompositor(layers...).Render()
}

func (m Model) detailsX() int {
	return max(m.width-m.detailsWidth(), 0)
}

func (m Model) detailsWidth() int {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	return m.detailsWidthForLayout(layout)
}

// offlineBadgeView renders the offline mode indicator.
func offlineBadgeView() string {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true).Padding(0, 1).Render("OFFLINE")
	}
	return offlineBadgeStyle.Render("OFFLINE")
}

func profileBadgeView(profile string, maxWidth int) string {
	profile = strings.TrimSpace(profile)
	if profile == "" || maxWidth <= 0 {
		return ""
	}

	padding := 1
	if maxWidth < 3 {
		padding = 0
	}
	profile = ansi.Truncate(profile, max(maxWidth-padding*2, 1), "…")
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true).Padding(0, padding).Render(profile)
	}
	return profileBadgeStyle.Padding(0, padding).Render(profile)
}

func (m Model) workspaceHeaderRightReserve() int {
	return lipgloss.Width(profileBadgeView(m.profileName, m.width))
}
