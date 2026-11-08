// Package textview presents a scrollable read-only report in the expanded popup.
package textview

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

type Model struct {
	title    string
	render   func(int) string
	viewport viewport.Model
}

func Open(width, height int, title string, render func(int) string) Model {
	v := viewport.New()
	v.KeyMap = viewport.KeyMap{
		Up:       tuiconfig.Binding(tuiconfig.BlockDialog, tuiconfig.ActionUp, "up"),
		Down:     tuiconfig.Binding(tuiconfig.BlockDialog, tuiconfig.ActionDown, "down"),
		PageUp:   tuiconfig.Binding(tuiconfig.BlockDialog, tuiconfig.ActionPageUp, "page up"),
		PageDown: tuiconfig.Binding(tuiconfig.BlockDialog, tuiconfig.ActionPageDown, "page down"),
	}
	m := Model{title: title, render: render, viewport: v}
	return m.SetSize(width, height)
}

func (m Model) IsOpen() bool { return m.render != nil }

func (m Model) Position() (int, int) {
	return viewutil.ExpandedPopupInset, viewutil.ExpandedPopupInset
}

func (m Model) SetSize(width, height int) Model {
	if m.IsOpen() {
		contentWidth := viewutil.ExpandedPopupContentWidth(width)
		m.viewport.SetWidth(contentWidth)
		m.viewport.SetHeight(viewutil.ExpandedPopupContentHeight(height))
		m.viewport.SetContentLines(strings.Split(m.render(contentWidth), "\n"))
	}
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch {
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionCancel, key.String()):
			return Model{}, nil
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionHome, key.String()):
			m.viewport.GotoTop()
			return m, nil
		case tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionEnd, key.String()):
			m.viewport.GotoBottom()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.IsOpen() {
		return ""
	}
	width, height := m.viewport.Width(), m.viewport.Height()
	inner := viewutil.PopupInnerWidth(width)
	border := styles.BorderStyle(true)
	side := border.Render("│")
	lines := []string{viewutil.PopupHeader(inner+2, "", m.title, true, border, "", styles.PanelMuted)}
	for range viewutil.PopupPaddingTop {
		lines = append(lines, side+viewutil.PopupContentLine("", width)+side)
	}
	scrollbar := viewutil.ScrollbarState(m.viewport.TotalLineCount(), m.viewport.YOffset(), height)
	for index, line := range strings.Split(m.viewport.View(), "\n") {
		edge := side
		if scrollbar.Visible && index >= scrollbar.ThumbStart && index <= scrollbar.ThumbEnd {
			edge = styles.ScrollbarThumb.Render("█")
		}
		lines = append(lines, side+viewutil.PopupContentLine(line, width)+edge)
	}
	footer := viewutil.ShortHelpView(width,
		viewutil.HelpBinding(m.viewport.KeyMap.Up.Help().Key+"/"+m.viewport.KeyMap.Down.Help().Key, "scroll"),
		viewutil.HelpBinding(m.viewport.KeyMap.PageUp.Help().Key+"/"+m.viewport.KeyMap.PageDown.Help().Key, "page"),
		tuiconfig.Binding(tuiconfig.BlockDialog, tuiconfig.ActionCancel, "close"),
	)
	lines = append(lines, side+viewutil.PopupContentLine(footer, width)+side)
	lines = append(lines, border.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return strings.Join(lines, "\n")
}
