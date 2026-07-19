package parameters

import (
	"context"
	"image/color"
	"reflect"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/messages"
)

type historyColumns struct {
	leftBorder, divider   int
	leftStart, rightStart int
	leftWidth, rightWidth int
}

var (
	historyAddedBackground   = lipgloss.Color("#315A46")
	historyRemovedBackground = lipgloss.Color("#68434A")
	historyChangedBackground = lipgloss.Color("#665A38")
)

func historyChangeBackground(kind diff.ChangeKind) color.Color {
	switch kind {
	case diff.ChangeAdded:
		return historyAddedBackground
	case diff.ChangeRemoved:
		return historyRemovedBackground
	case diff.ChangeChanged:
		return historyChangedBackground
	default:
		return nil
	}
}

func (m Model) historyStacked() bool {
	layout := m.parameterRenderLayout()
	if layout.mode == parameterRenderModeNarrow {
		return true
	}
	columns := m.historyColumnLayout()
	return columns.leftWidth < 12 || columns.rightWidth < 12
}

func (m Model) historyColumnLayout() historyColumns {
	width := m.viewportWidth()
	leftBorder := min(max(m.parameterRenderLayout().valueStart-2, 1), max(width-3, 1))
	divider := leftBorder + max((width-leftBorder)/2, 1)
	divider = min(divider, max(width-2, leftBorder+1))
	return historyColumns{
		leftBorder: leftBorder, divider: divider,
		leftStart: leftBorder + 1, rightStart: divider + 1,
		leftWidth: max(divider-leftBorder-1, 0), rightWidth: max(width-divider-1, 0),
	}
}

func (m Model) renderHistoryGridLine(line string, selected bool, kind visibleNodeKind) string {
	width := m.viewportWidth()
	line = viewutil.PadRight(ansi.Truncate(line, width, ""), width)
	fullSelection := selected && kind != nodeValue
	if fullSelection && kind == nodeProject {
		return line
	}
	if fullSelection {
		selection := parameterSelectionStyle()
		if kind == nodeGroup {
			selection = groupSelectionStyle()
		}
		line = selection.Render(ansi.Strip(line))
	}
	return line
}

func (m Model) historyParameterPair(projectID, groupKey, paramKey string) (*core.ParametersEntry, *core.ParametersEntry, diff.ChangeKind) {
	state := m.histories[projectID]
	key := historyParamKey(groupKey, paramKey)
	previous := state.previousParams[key]
	current := state.currentParams[key]
	return previous, current, state.paramKinds[key]
}

func (m Model) historyTree(projectID string, fallback *core.ParametersTree) *core.ParametersTree {
	state := m.histories[projectID]
	if state.merged == nil {
		return fallback
	}
	return state.merged
}

func buildHistoryState(state historyState) historyState {
	if state.current == nil {
		return state
	}
	merged := *state.current
	merged.Groups = append([]core.ParametersGroup(nil), state.current.Groups...)
	for gi := range merged.Groups {
		merged.Groups[gi].Parameters = append([]core.ParametersEntry(nil), merged.Groups[gi].Parameters...)
	}
	previousGroups := []core.ParametersGroup(nil)
	if state.previous != nil {
		previousGroups = state.previous.Groups
	}
	for _, oldGroup := range previousGroups {
		gi := -1
		for i := range merged.Groups {
			if merged.Groups[i].Key == oldGroup.Key {
				gi = i
				break
			}
		}
		if gi < 0 {
			merged.Groups = append(merged.Groups, oldGroup)
			continue
		}
		for _, oldParam := range oldGroup.Parameters {
			found := false
			for pi := range merged.Groups[gi].Parameters {
				if merged.Groups[gi].Parameters[pi].Key == oldParam.Key {
					found = true
					for _, oldValue := range oldParam.Values {
						valueFound := false
						for _, value := range merged.Groups[gi].Parameters[pi].Values {
							if value.Label == oldValue.Label {
								valueFound = true
								break
							}
						}
						if !valueFound {
							merged.Groups[gi].Parameters[pi].Values = append(merged.Groups[gi].Parameters[pi].Values, oldValue)
						}
					}
					break
				}
			}
			if !found {
				merged.Groups[gi].Parameters = append(merged.Groups[gi].Parameters, oldParam)
			}
		}
	}
	for gi := range merged.Groups {
		slices.SortFunc(merged.Groups[gi].Parameters, func(left, right core.ParametersEntry) int {
			return strfold.Compare(left.Key, right.Key)
		})
	}
	state.merged = &merged
	state.previousParams, state.previousValues = indexHistoryTree(state.previous)
	state.currentParams, state.currentValues = indexHistoryTree(state.current)
	state.mergedParams, state.mergedValues = indexHistoryTree(state.merged)
	state.paramKinds = make(map[string]diff.ChangeKind, len(state.mergedParams))
	state.valueKinds = make(map[string]map[string]diff.ChangeKind, len(state.mergedValues))
	state.counts = historyChangeCounts{}
	for key := range state.mergedParams {
		state.paramKinds[key] = compareHistoryItems(state.previousParams[key], state.currentParams[key])
		switch state.paramKinds[key] {
		case diff.ChangeAdded:
			state.counts.added++
		case diff.ChangeRemoved:
			state.counts.removed++
		case diff.ChangeChanged:
			state.counts.changed++
		}
		byLabel := make(map[string]diff.ChangeKind, len(state.mergedValues[key]))
		for label := range state.mergedValues[key] {
			byLabel[label] = compareHistoryItems(state.previousValues[key][label], state.currentValues[key][label])
		}
		state.valueKinds[key] = byLabel
	}
	return state
}

func compareHistoryItems[T any](previous, current *T) diff.ChangeKind {
	switch {
	case previous == nil && current != nil:
		return diff.ChangeAdded
	case previous != nil && current == nil:
		return diff.ChangeRemoved
	case reflect.DeepEqual(previous, current):
		return diff.ChangeUnchanged
	default:
		return diff.ChangeChanged
	}
}

func historyParamKey(groupKey, paramKey string) string { return groupKey + "\x00" + paramKey }

func indexHistoryTree(tree *core.ParametersTree) (map[string]*core.ParametersEntry, map[string]map[string]*core.ParametersValue) {
	params := make(map[string]*core.ParametersEntry)
	values := make(map[string]map[string]*core.ParametersValue)
	if tree == nil {
		return params, values
	}
	for gi := range tree.Groups {
		group := &tree.Groups[gi]
		for pi := range group.Parameters {
			param := &group.Parameters[pi]
			key := historyParamKey(group.Key, param.Key)
			params[key] = param
			byLabel := make(map[string]*core.ParametersValue, len(param.Values))
			for vi := range param.Values {
				byLabel[param.Values[vi].Label] = &param.Values[vi]
			}
			values[key] = byLabel
		}
	}
	return params, values
}

func (m Model) historyValuePair(projectID, groupKey, paramKey, label string) (*core.ParametersValue, *core.ParametersValue) {
	state := m.histories[projectID]
	key := historyParamKey(groupKey, paramKey)
	return state.previousValues[key][label], state.currentValues[key][label]
}

func (m Model) historyValueKind(projectID, groupKey, paramKey, label string) diff.ChangeKind {
	state := m.histories[projectID]
	return state.valueKinds[historyParamKey(groupKey, paramKey)][label]
}

func (m Model) historyMergedParameter(projectID, groupKey, paramKey string) *core.ParametersEntry {
	return m.histories[projectID].mergedParams[historyParamKey(groupKey, paramKey)]
}

func (m Model) historyMergedValue(projectID, groupKey, paramKey, label string) *core.ParametersValue {
	state := m.histories[projectID]
	return state.mergedValues[historyParamKey(groupKey, paramKey)][label]
}

func historyValueText(param *core.ParametersEntry, width int) string {
	if param == nil || width <= 0 {
		return ""
	}
	var out strings.Builder
	remaining := width
	for i, value := range param.Values {
		if i > 0 {
			if remaining <= 3 {
				break
			}
			out.WriteString(" / ")
			remaining -= 3
		}
		part := ansi.Truncate(value.Value, remaining, "")
		out.WriteString(part)
		remaining -= lipgloss.Width(part)
		if remaining <= 0 {
			break
		}
	}
	return out.String()
}

func historyParameterIcon(param *core.ParametersEntry) string {
	if param == nil {
		return ""
	}
	if param != nil && len(param.Values) > 1 {
		return "⌥"
	}
	return "╌"
}

func (m Model) loadHistoryCmd(project core.Project, preferred historyPairSelection, hasPreferred bool) tea.Cmd {
	return func() tea.Msg {
		list, err := m.svc.ListRemoteConfigVersions(context.Background(), project.ProjectID, core.VersionListOptions{All: true})
		if err != nil {
			return messages.HistoryLoadedMsg{Project: project, Err: err}
		}
		if len(list.Versions) < 2 {
			return messages.HistoryLoadedMsg{Project: project, Versions: list.Versions, Unavailable: true}
		}
		if hasPreferred && historyVersionIndex(list.Versions, preferred.previous) >= 0 && historyVersionIndex(list.Versions, preferred.current) >= 0 {
			return m.loadHistoryPair(project, preferred.previous, preferred.current, list.Versions)
		}
		return m.loadHistoryPair(project, list.Versions[1].VersionNumber, list.Versions[0].VersionNumber, list.Versions)
	}
}

func (m Model) loadHistoryPairCmd(project core.Project, previousVersion, currentVersion string) tea.Cmd {
	return func() tea.Msg { return m.loadHistoryPair(project, previousVersion, currentVersion, nil) }
}

func (m Model) loadHistoryPair(project core.Project, previousVersion, currentVersion string, versions []core.RemoteConfigVersionEntry) messages.HistoryLoadedMsg {
	previous, current, err := m.svc.GetRemoteConfigVersionPair(context.Background(), project.ProjectID, previousVersion, currentVersion, false)
	if err != nil {
		return messages.HistoryLoadedMsg{Project: project, Versions: versions, Err: err}
	}
	previousTree, err := m.svc.BuildParametersTree(previous.Cache)
	if err != nil {
		return messages.HistoryLoadedMsg{Project: project, Versions: versions, Err: err}
	}
	currentTree, err := m.svc.BuildParametersTree(current.Cache)
	if err != nil {
		return messages.HistoryLoadedMsg{Project: project, Versions: versions, Err: err}
	}
	return messages.HistoryLoadedMsg{Project: project, PreviousTree: previousTree, CurrentTree: currentTree,
		PreviousVersion: previous.Version.VersionNumber, CurrentVersion: current.Version.VersionNumber,
		PreviousPublished: formatPublished(previous.Version.UpdateTime), CurrentPublished: formatPublished(current.Version.UpdateTime), Versions: versions}
}

func formatPublished(raw string) string {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return raw
	}
	return parsed.Local().Format("2006-01-02 15:04:05")
}

func (m *Model) updateHistory(msg messages.HistoryLoadedMsg) {
	old := m.histories[msg.Project.ProjectID]
	versions := msg.Versions
	if len(versions) == 0 {
		versions = old.versions
	}
	pairs := old.pairs
	if pairs == nil {
		pairs = make(map[string]historyPairData)
	}
	state := historyState{
		previous: msg.PreviousTree, current: msg.CurrentTree,
		previousVersion: msg.PreviousVersion, currentVersion: msg.CurrentVersion,
		previousPublished: msg.PreviousPublished, currentPublished: msg.CurrentPublished, err: msg.Err,
		unavailable: msg.Unavailable, versions: versions, pairs: pairs,
	}
	if state.unavailable && state.currentVersion == "" && len(versions) > 0 {
		state.currentVersion = versions[0].VersionNumber
	}
	if msg.Err != nil && old.current != nil {
		old.err, old.loading = msg.Err, false
		old.versions = versions
		m.histories[msg.Project.ProjectID] = old
		return
	}
	state = buildHistoryState(state)
	state.pairs[historyPairKey(state.previousVersion, state.currentVersion)] = historyPairData{previous: state.previous, current: state.current,
		previousVersion: state.previousVersion, currentVersion: state.currentVersion, previousPublished: state.previousPublished, currentPublished: state.currentPublished}
	m.histories[msg.Project.ProjectID] = state
	m.syncVisible()
}

func historyPairKey(left, right string) string { return left + "\x00" + right }

func (m *Model) invalidateHistoryIfVersionChanged(projectID string) {
	history, ok := m.histories[projectID]
	if !ok || history.loading {
		return
	}
	idx, ok := m.projectIndex[projectID]
	if !ok {
		return
	}
	project := m.projects[idx]
	desired := project.cacheVersion
	if desired == "" && project.tree != nil {
		desired = project.tree.Version
	}
	if history.unavailable {
		if desired == "" || desired == history.currentVersion {
			return
		}
		delete(m.histories, projectID)
		m.syncVisible()
		return
	}
	if len(history.versions) == 0 {
		return
	}
	if desired == "" || desired == history.versions[0].VersionNumber {
		return
	}
	delete(m.histories, projectID)
	m.syncVisible()
}
