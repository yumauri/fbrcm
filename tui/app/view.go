package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/minsize"
	"github.com/yumauri/fbrcm/tui/panels"
)

var rootStyle = lipgloss.NewStyle()

// View handles view for Model and returns the resulting state or error.
func (m Model) View() tea.View {
	if m.width < minsize.MinWidth || m.height < minsize.MinHeight {
		v := tea.NewView(rootStyle.Render(minsize.View(m.width, m.height)))
		v.AltScreen = true
		v.MouseMode = tea.MouseModeNone
		return v
	}

	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.projects.View(m.active == panels.Projects),
		m.parameters.View(m.active == panels.Parameters),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		topRow,
		m.logs.View(m.active == panels.Logs),
		m.helpView(),
	)

	layers := []*lipgloss.Layer{lipgloss.NewLayer(body).ID("base")}
	if m.detailsVisible {
		layers = append(layers, lipgloss.NewLayer(m.details.View()).ID("details").X(m.detailsX()).Y(0).Z(1))
		if m.details.DropdownOpen() {
			x, y := m.details.DropdownCurrentPosition()
			layers = append(layers, lipgloss.NewLayer(m.details.DropdownCurrentView()).ID("details-dropdown-current").X(x).Y(y).Z(2))
			x, y = m.details.DropdownListPosition()
			layers = append(layers, lipgloss.NewLayer(m.details.DropdownListView()).ID("details-dropdown-list").X(x).Y(y).Z(2))
		}
	}
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
	if m.renameInput.IsOpen() {
		x, y := m.renameInput.Position()
		layers = append(layers, lipgloss.NewLayer(m.renameInput.View()).ID("rename").X(x).Y(y).Z(3))
	}
	if m.dialog.IsOpen() {
		dialog := m.dialog.View()
		x, y := m.dialog.Position()
		layers = append(layers, lipgloss.NewLayer(dialog).ID("dialog").X(x).Y(y).Z(4))
	}
	if len(layers) > 1 {
		body = lipgloss.NewCompositor(layers...).Render()
	}

	v := tea.NewView(rootStyle.Render(body))
	v.AltScreen = true
	if m.active == panels.Logs {
		v.MouseMode = tea.MouseModeNone
	} else {
		v.MouseMode = tea.MouseModeAllMotion
	}
	return v
}

// detailsX handles details x for Model and returns the resulting state or error.
func (m Model) detailsX() int {
	return max(m.width-m.detailsWidth(), 0)
}

// detailsWidth handles details width for Model and returns the resulting state or error.
func (m Model) detailsWidth() int {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	return m.detailsWidthForLayout(layout)
}
