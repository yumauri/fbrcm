package promote

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.IsOpen() {
		return m, nil
	}
	switch mouse := msg.(type) {
	case tea.MouseClickMsg:
		return m.updateMouseClick(mouse)
	case tea.MouseWheelMsg:
		if m.TargetPickerOpen() {
			return m.updateTargetMouseWheel(mouse)
		}
		if mouse.Mouse().X >= m.x && mouse.Mouse().X < m.x+m.width && mouse.Mouse().Y >= m.y && mouse.Mouse().Y < m.y+m.height {
			delta := 1
			if mouse.Mouse().Button == tea.MouseWheelUp {
				delta = -1
			}
			m.move(delta)
			m.syncDetail()
		}
		return m, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if m.TargetPickerOpen() {
		if !ok {
			return m.updateTargetInput(msg)
		}
		k := key.String()
		if tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionClose, k) {
			if m.pickerOpen {
				return m.cancelTargetPicker(), nil
			}
			return m, cmd(CloseRequestedMsg{})
		}
		if m.targetNavigationKey(k) {
			return m.updateTargetKey(k)
		}
		return m.updateTargetInput(msg)
	}
	if !ok {
		if m.filter.Focused() {
			return m.updateFilter(msg)
		}
		return m, nil
	}
	k := key.String()
	if m.filter.Focused() {
		return m.updateFilterKey(key, k)
	}
	if mode, ok := tuiconfig.FilterModeForKey(k); ok {
		return m, m.filter.Activate(mode)
	}
	if tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionClose, k) {
		return m, cmd(CloseRequestedMsg{})
	}
	if m.phase != phaseReview || m.loading {
		return m, nil
	}
	return m.updateReviewKey(k)
}

func (m Model) updateMouseClick(msg tea.MouseClickMsg) (Model, tea.Cmd) {
	mouse := msg.Mouse()
	if mouse.Button != tea.MouseLeft {
		return m, nil
	}
	if m.TargetPickerOpen() {
		x, y := m.TargetPosition()
		view := m.TargetView()
		if mouse.X < x || mouse.X >= x+lipgloss.Width(view) || mouse.Y < y || mouse.Y >= y+lipgloss.Height(view) {
			return m, nil
		}
		row := mouse.Y - y - targetFirstOptionRow
		index := row
		if row < 0 || index < 0 || index >= len(m.candidates) {
			return m, nil
		}
		double := m.lastClick.index == index && time.Since(m.lastClick.at) <= 400*time.Millisecond
		m.pickerCursor = index
		m.lastClick.index, m.lastClick.at = index, time.Now()
		if double {
			return m, cmd(TargetSelectedMsg{Source: m.pickerSource, Target: m.candidates[index], Mode: m.sourceMode})
		}
		return m, nil
	}
	if mouse.X < m.x || mouse.X >= m.x+m.width || mouse.Y < m.y || mouse.Y >= m.y+m.height {
		return m, nil
	}
	if m.phase != phaseReview {
		return m, nil
	}
	innerWidth := max(m.width-2, 1)
	leftWidth, _ := m.promotionColumnWidths(innerWidth)
	if mouse.X-m.x >= leftWidth {
		return m, nil
	}
	bodyRow := mouse.Y - m.y - 1 - promoteHeaderHeight
	index, ok := m.itemIndexAtBodyRow(bodyRow)
	if !ok {
		return m, nil
	}
	double := m.lastClick.index == index && time.Since(m.lastClick.at) <= 400*time.Millisecond
	m.cursor = index
	m.lastClick.index, m.lastClick.at = index, time.Now()
	m.ensureVisible()
	m.syncDetail()
	if double {
		m = m.toggleCurrent()
	}
	return m, nil
}

func (m Model) updateTargetMouseWheel(msg tea.MouseWheelMsg) (Model, tea.Cmd) {
	mouse := msg.Mouse()
	x, y := m.TargetPosition()
	view := m.TargetView()
	if mouse.X < x || mouse.X >= x+lipgloss.Width(view) || mouse.Y < y || mouse.Y >= y+lipgloss.Height(view) {
		return m, nil
	}
	delta := 1
	if mouse.Button == tea.MouseWheelUp {
		delta = -1
	}
	m.moveTarget(delta)
	return m, nil
}

func (m Model) itemIndexAtBodyRow(row int) (int, bool) {
	if row < 2 {
		return 0, false
	}
	physical := 2 // "Changes" heading and its following empty row.
	lastKind := rcdiff.ItemKind("")
	for index := m.offset; index < len(m.visible); index++ {
		item := m.visible[index]
		if item.Kind != lastKind {
			if lastKind != "" {
				physical++
			}
			physical++
			lastKind = item.Kind
		}
		if physical == row {
			return index, true
		}
		physical++
	}
	return 0, false
}

func (m Model) updateTargetKey(k string) (Model, tea.Cmd) {
	switch {
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionUp, k):
		m.moveTarget(-1)
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionDown, k):
		m.moveTarget(1)
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPageUp, k):
		m.moveTarget(-m.pageSize())
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPageDown, k):
		m.moveTarget(m.pageSize())
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionHome, k):
		m.pickerCursor = 0
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionEnd, k):
		m.pickerCursor = max(len(m.candidates)-1, 0)
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSubmit, k):
		if m.pickerCursor >= 0 && m.pickerCursor < len(m.candidates) {
			return m, cmd(TargetSelectedMsg{Source: m.pickerSource, Target: m.candidates[m.pickerCursor], Mode: m.sourceMode})
		}
	}
	return m, nil
}

func (m Model) targetNavigationKey(k string) bool {
	return tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionUp, k) ||
		tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionDown, k) ||
		tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPageUp, k) ||
		tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPageDown, k) ||
		tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionHome, k) ||
		tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionEnd, k) ||
		tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSubmit, k)
}

func (m Model) updateTargetInput(msg tea.Msg) (Model, tea.Cmd) {
	var command tea.Cmd
	m.targetInput, command = m.targetInput.Update(msg)
	m.applyProjectFilter()
	return m, command
}

func (m Model) updateReviewKey(k string) (Model, tea.Cmd) {
	switch {
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionUp, k):
		m.move(-1)
		m.syncDetail()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionDown, k):
		m.move(1)
		m.syncDetail()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPageUp, k):
		m.move(-m.pageSize())
		m.syncDetail()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPageDown, k):
		m.move(m.pageSize())
		m.syncDetail()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionHome, k):
		m.cursor = 0
		m.ensureVisible()
		m.syncDetail()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionEnd, k):
		m.cursor = max(len(m.visible)-1, 0)
		m.ensureVisible()
		m.syncDetail()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionToggle, k):
		m = m.toggleCurrent()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSubmit, k):
		if item, ok := m.CurrentItem(); ok {
			return m, cmd(DiffRequestedMsg{Item: item})
		}
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSelectAll, k):
		for _, item := range m.visible {
			if item.Change != rcdiff.ChangeRemoved || m.prune {
				m.requested[item.ID] = true
			}
		}
		m = m.rebuildPreview()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSelectNone, k):
		for _, item := range m.visible {
			delete(m.requested, item.ID)
		}
		m = m.rebuildPreview()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPrune, k):
		m.prune = !m.prune
		if !m.prune {
			for _, item := range m.plan.Plan.Items {
				if item.Change == rcdiff.ChangeRemoved {
					delete(m.requested, item.ID)
				}
			}
		}
		m = m.rebuildPreview()
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSwap, k):
		return m, cmd(SwapRequestedMsg{})
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSource, k):
		if m.plan.Source.HasDraft {
			mode := core.ProjectPromotionPublished
			if m.sourceMode == core.ProjectPromotionPublished {
				mode = core.ProjectPromotionEffective
			}
			return m, cmd(SourceModeRequestedMsg{Source: m.source, Target: m.target, Mode: mode})
		}
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionSaveDraft, k):
		if m.HasSelection() {
			return m, cmd(SaveRequestedMsg{Preview: m.preview})
		}
	case tuiconfig.Matches(tuiconfig.BlockPromote, tuiconfig.ActionPublish, k):
		if m.CanPublish() {
			return m, cmd(PublishRequestedMsg{Preview: m.preview})
		}
	}
	return m, nil
}

func (m Model) toggleCurrent() Model {
	item, ok := m.CurrentItem()
	if !ok || !m.changeSelectable(item) {
		return m
	}
	if m.requested[item.ID] {
		delete(m.requested, item.ID)
	} else {
		m.requested[item.ID] = true
	}
	return m.rebuildPreview()
}

func (m Model) updateFilterKey(msg tea.KeyMsg, k string) (Model, tea.Cmd) {
	switch {
	case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterApply, k):
		m.filter.Blur()
		return m, nil
	case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterCancel, k):
		m.filter.ClearAndBlur()
		m.applyFilter()
		return m, nil
	case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterUp, k):
		m.filter.Blur()
		m.move(-1)
		return m, nil
	case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterDown, k):
		m.filter.Blur()
		m.move(1)
		return m, nil
	}
	return m.updateFilter(msg)
}

func (m Model) updateFilter(msg tea.Msg) (Model, tea.Cmd) {
	var command tea.Cmd
	m.filter, command = m.filter.Update(msg)
	m.applyFilter()
	return m, command
}

func (m *Model) applyFilter() {
	m.applyItemFilter()
	m.syncDetail()
}

func (m *Model) move(delta int) {
	length := len(m.visible)
	if length == 0 {
		m.cursor = 0
		return
	}
	m.cursor = max(0, min(m.cursor+delta, length-1))
	m.ensureVisible()
}

func (m *Model) moveTarget(delta int) {
	if len(m.candidates) == 0 {
		m.pickerCursor = 0
		return
	}
	m.pickerCursor = max(0, min(m.pickerCursor+delta, len(m.candidates)-1))
}

func (m Model) pageSize() int { return max(m.bodyHeight()-7, 1) }

func (m Model) bodyHeight() int {
	return max(m.height-2-promoteHeaderHeight-m.filter.Height(), 1)
}

func (m *Model) ensureVisible() {
	page := m.pageSize()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+page {
		m.offset = m.cursor - page + 1
	}
	m.offset = max(m.offset, 0)
}
