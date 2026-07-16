package conditions

import "github.com/yumauri/fbrcm/core"

// StartMove begins a reversible reorder operation for the selected condition.
func (m *Model) StartMove() bool {
	data, ok := m.currentData()
	if !ok || m.move != nil {
		return false
	}
	projectIndex, ok := m.projectIndex[data.Project.ProjectID]
	if !ok || m.projects[projectIndex].tree == nil || len(m.projects[projectIndex].tree.Conditions) < 2 {
		return false
	}
	m.move = &conditionMoveState{
		projectID:     data.Project.ProjectID,
		conditionName: data.Condition.Name,
		original:      append([]core.ConditionEntry(nil), m.projects[projectIndex].tree.Conditions...),
	}
	return true
}

// MoveActive reports whether a condition reorder operation is active.
func (m Model) MoveActive() bool {
	return m.move != nil
}

// MoveActiveCondition inserts the moving condition one position up or down.
// Movement is confined to the condition's original project.
func (m *Model) MoveActiveCondition(delta int) bool {
	if m.move == nil || delta == 0 {
		return false
	}
	projectIndex, ok := m.projectIndex[m.move.projectID]
	if !ok || m.projects[projectIndex].tree == nil {
		return false
	}
	conditions := append([]core.ConditionEntry(nil), m.projects[projectIndex].tree.Conditions...)
	current := conditionIndexByName(conditions, m.move.conditionName)
	if current < 0 {
		return false
	}
	direction := 1
	if delta < 0 {
		direction = -1
	}
	next := current + direction
	if next < 0 || next >= len(conditions) {
		return false
	}
	conditions[current], conditions[next] = conditions[next], conditions[current]
	setConditionPriorities(conditions)
	m.setProjectConditions(projectIndex, conditions)
	m.syncVisible()
	m.selectCondition(m.move.projectID, m.move.conditionName)
	return true
}

// CancelMove restores the condition order from before move mode started.
func (m *Model) CancelMove() {
	if m.move == nil {
		return
	}
	projectID, conditionName := m.move.projectID, m.move.conditionName
	projectIndex, ok := m.projectIndex[projectID]
	if ok {
		m.setProjectConditions(projectIndex, m.move.original)
	}
	m.move = nil
	m.syncVisible()
	m.selectCondition(projectID, conditionName)
}

// FinishMove returns the requested priority and restores the loaded tree so
// the normal preview/draft mutation pipeline remains the source of truth.
func (m *Model) FinishMove() (priority int, changed bool, ok bool) {
	if m.move == nil {
		return 0, false, false
	}
	projectID, conditionName := m.move.projectID, m.move.conditionName
	projectIndex, exists := m.projectIndex[projectID]
	if !exists || m.projects[projectIndex].tree == nil {
		m.move = nil
		return 0, false, false
	}
	current := conditionIndexByName(m.projects[projectIndex].tree.Conditions, conditionName)
	original := conditionIndexByName(m.move.original, conditionName)
	if current < 0 || original < 0 {
		m.CancelMove()
		return 0, false, false
	}
	priority = current + 1
	changed = current != original
	m.CancelMove()
	return priority, changed, true
}

func (m Model) movingCondition(projectID, conditionName string) bool {
	return m.move != nil && m.move.projectID == projectID && m.move.conditionName == conditionName
}

func (m *Model) setProjectConditions(projectIndex int, conditions []core.ConditionEntry) {
	if projectIndex < 0 || projectIndex >= len(m.projects) || m.projects[projectIndex].tree == nil {
		return
	}
	project := m.projects[projectIndex]
	tree := *project.tree
	tree.Conditions = append([]core.ConditionEntry(nil), conditions...)
	project.tree = &tree
	m.projects[projectIndex] = project
}

func conditionIndexByName(conditions []core.ConditionEntry, name string) int {
	for index := range conditions {
		if conditions[index].Name == name {
			return index
		}
	}
	return -1
}

func setConditionPriorities(conditions []core.ConditionEntry) {
	for index := range conditions {
		conditions[index].Priority = index + 1
	}
}
