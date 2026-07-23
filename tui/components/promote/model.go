package promote

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corefilter "github.com/yumauri/fbrcm/core/filter"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
	"github.com/yumauri/fbrcm/tui/components/inputstyles"
)

type phase int

const (
	phaseClosed phase = iota
	phaseTarget
	phaseLoading
	phaseReview
)

type TargetSelectedMsg struct {
	Source core.Project
	Target core.Project
	Mode   core.ProjectPromotionSourceMode
}

type CloseRequestedMsg struct{}
type SwapRequestedMsg struct{}
type DiffRequestedMsg struct{ Item rcpromote.Item }
type SourceModeRequestedMsg struct {
	Source core.Project
	Target core.Project
	Mode   core.ProjectPromotionSourceMode
}
type SaveRequestedMsg struct{ Preview *core.ProjectPromotionPreview }
type PublishRequestedMsg struct{ Preview *core.ProjectPromotionPreview }

type Model struct {
	svc *core.Core

	phase        phase
	source       core.Project
	target       core.Project
	projects     []core.Project
	candidates   []core.Project
	pickerOpen   bool
	pickerSource core.Project
	pickerCursor int
	plan         *core.ProjectPromotionPlan
	preview      *core.ProjectPromotionPreview
	requested    map[rcpromote.ItemID]bool
	visible      []rcpromote.Item
	prune        bool
	verified     bool
	loading      bool
	err          error
	sourceMode   core.ProjectPromotionSourceMode
	cursor       int
	offset       int
	targetInput  textinput.Model
	targetRow    int
	filter       filterbox.Model
	detail       viewport.Model
	x, y         int
	width        int
	height       int
	lastClick    struct {
		index int
		at    time.Time
	}
}

func New(svc *core.Core) Model {
	detail := viewport.New(viewport.WithWidth(1), viewport.WithHeight(1))
	detail.SoftWrap = false
	targetInput := inputstyles.NewTextInput()
	targetInput.Prompt = "Filter: "
	targetInput.Placeholder = "Type to filter projects"
	m := Model{svc: svc, requested: make(map[rcpromote.ItemID]bool), targetInput: targetInput, filter: filterbox.New(), detail: detail}
	m.lastClick.index = -1
	return m
}

func (m Model) IsOpen() bool                           { return m.phase != phaseClosed }
func (m Model) IsReviewing() bool                      { return m.phase == phaseReview }
func (m Model) TargetPickerOpen() bool                 { return m.phase == phaseTarget || m.pickerOpen }
func (m Model) WorkspaceOpen() bool                    { return m.phase == phaseLoading || m.phase == phaseReview }
func (m Model) FilterFocused() bool                    { return m.WorkspaceOpen() && m.filter.Focused() }
func (m Model) Source() core.Project                   { return m.source }
func (m Model) Target() core.Project                   { return m.target }
func (m Model) Preview() *core.ProjectPromotionPreview { return m.preview }

func (m Model) Open(source core.Project, projects []core.Project) Model {
	m.phase = phaseTarget
	m.source = source
	m.target = core.Project{}
	m.plan = nil
	m.preview = nil
	m.requested = make(map[rcpromote.ItemID]bool)
	m.prune = false
	m.verified = false
	m.loading = false
	m.err = nil
	m.sourceMode = core.ProjectPromotionEffective
	m.cursor = 0
	m.offset = 0
	m.openTargetPicker(source, projects)
	m.filter.ClearAndBlur()
	return m
}

// OpenTargetPicker opens a new source/target selection without discarding an
// existing promotion workspace. Initial promotion selection still starts from
// a clean model through Open.
func (m Model) OpenTargetPicker(source core.Project, projects []core.Project) Model {
	if !m.WorkspaceOpen() {
		return m.Open(source, projects)
	}
	m.openTargetPicker(source, projects)
	return m
}

func (m *Model) openTargetPicker(source core.Project, projects []core.Project) {
	m.pickerOpen = m.phase != phaseTarget
	m.pickerSource = source
	m.pickerCursor = 0
	m.projects = append([]core.Project(nil), projects...)
	m.targetInput.SetValue("")
	m.targetInput.CursorEnd()
	_ = m.targetInput.Focus()
	m.applyProjectFilter()
}

func (m Model) cancelTargetPicker() Model {
	m.pickerOpen = false
	m.pickerSource = core.Project{}
	m.projects = nil
	m.candidates = nil
	m.pickerCursor = 0
	m.targetInput.Blur()
	return m
}

func (m Model) Close() Model {
	return New(m.svc)
}

func (m Model) SetBounds(x, y, width, height int) Model {
	m.x, m.y, m.width, m.height = x, y, width, height
	m.syncLayout()
	return m
}

// SetTargetRow anchors the selected target option to the source project's
// rendered row while the target list moves around it.
func (m Model) SetTargetRow(row int) Model {
	m.targetRow = row
	return m
}

func (m Model) SetLoading(source, target core.Project, mode core.ProjectPromotionSourceMode) Model {
	m.phase = phaseLoading
	m.source, m.target, m.sourceMode = source, target, mode
	m.pickerOpen = false
	m.pickerSource = core.Project{}
	m.projects = nil
	m.candidates = nil
	m.pickerCursor = 0
	m.targetInput.Blur()
	m.loading = true
	m.err = nil
	m.plan = nil
	m.preview = nil
	return m
}

func (m Model) SetPlan(plan *core.ProjectPromotionPlan, verified bool) Model {
	m.phase = phaseReview
	m.loading = false
	m.err = nil
	m.plan = plan
	m.verified = verified
	m.requested = make(map[rcpromote.ItemID]bool)
	m.cursor, m.offset = 0, 0
	m.applyItemFilter()
	return m.rebuildPreview()
}

func (m Model) SetError(err error) Model {
	m.loading = false
	m.err = err
	if m.plan == nil {
		m.phase = phaseReview
	}
	return m
}

func (m Model) SetSaved(plan *core.ProjectPromotionPlan) Model {
	return m.SetPlan(plan, m.verified)
}

func (m *Model) applyProjectFilter() {
	selectedID := ""
	if m.pickerCursor >= 0 && m.pickerCursor < len(m.candidates) {
		selectedID = m.candidates[m.pickerCursor].ProjectID
	}
	query := strings.ToLower(strings.TrimSpace(m.targetInput.Value()))
	m.candidates = m.candidates[:0]
	for _, project := range m.projects {
		if project.ProjectID == m.pickerSource.ProjectID {
			continue
		}
		value := strings.ToLower(project.Name + " " + project.ProjectID)
		if strings.Contains(value, query) {
			m.candidates = append(m.candidates, project)
		}
	}
	m.pickerCursor = 0
	if selectedID != "" {
		for index, project := range m.candidates {
			if project.ProjectID == selectedID {
				m.pickerCursor = index
				break
			}
		}
	}
}

func (m *Model) applyItemFilter() {
	m.visible = m.visible[:0]
	if m.plan == nil {
		return
	}
	for _, item := range m.plan.Plan.Items {
		value := string(item.Kind) + " " + string(item.Change) + " " + item.Label + " " + item.ID.Group
		matched, _ := corefilter.Match(value, m.filter.Value(), m.filter.Mode())
		if matched {
			m.visible = append(m.visible, item)
		}
	}
	m.cursor = max(0, min(m.cursor, len(m.visible)-1))
	m.ensureVisible()
}

func (m Model) rebuildPreview() Model {
	if m.svc == nil || m.plan == nil {
		return m
	}
	preview, err := m.svc.PreviewProjectPromotion(m.plan, m.requested, m.prune)
	m.preview, m.err = preview, err
	m.syncDetail()
	return m
}

func (m Model) CurrentItem() (rcpromote.Item, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return rcpromote.Item{}, false
	}
	return m.visible[m.cursor], true
}

func (m Model) HasSelection() bool {
	return m.preview != nil && m.preview.HasChanges && len(m.preview.Requested) > 0
}

func (m Model) CanPublish() bool { return m.HasSelection() && m.verified && !m.target.Disabled }

func cmd(msg tea.Msg) tea.Cmd { return func() tea.Msg { return msg } }
