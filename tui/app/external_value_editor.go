package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	coreeditor "github.com/yumauri/fbrcm/core/editor"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	jsoninput "github.com/yumauri/fbrcm/tui/components/jsoninput"
	"github.com/yumauri/fbrcm/tui/panels"
)

const (
	// Bubbles textarea silently truncates input beyond 10,000 logical lines.
	// Keep this guard even if the lower performance threshold changes.
	textareaHardLineLimit = 10_000
	builtInEditorMaxLines = 1_000
	builtInEditorMaxBytes = 128 * 1024
)

type externalValueKind uint8

const (
	externalStringValue externalValueKind = iota + 1
	externalJSONValue
)

type externalValueEditSession struct {
	kind         externalValueKind
	project      core.Project
	groupKey     string
	paramKey     string
	valueLabel   string
	currentValue string
	source       panels.ID
	path         string
}

type externalValueEditorFinishedMsg struct {
	path string
	err  error
}

type externalValueEditorReopenMsg struct{}
type externalValueEditorDismissedMsg struct{}

func shouldUseExternalValueEditor(value string) bool {
	lineCount := strings.Count(value, "\n") + 1
	if lineCount > textareaHardLineLimit {
		return true
	}
	return lineCount > builtInEditorMaxLines || len(value) > builtInEditorMaxBytes
}

func (m *Model) openExternalJSONValueEditor() tea.Cmd {
	anchor, ok := m.currentJSONValueAnchor()
	if !ok {
		return nil
	}
	return m.openExternalValueEditor(externalValueEditSession{
		kind:         externalJSONValue,
		project:      anchor.Project,
		groupKey:     anchor.GroupKey,
		paramKey:     anchor.ParamKey,
		valueLabel:   anchor.ValueLabel,
		currentValue: anchor.CurrentValue,
		source:       m.currentValueEditSource(),
	})
}

func (m *Model) openExternalStringValueEditor() tea.Cmd {
	anchor, ok := m.currentStringValueAnchor()
	if !ok {
		return nil
	}
	return m.openExternalValueEditor(externalValueEditSession{
		kind:         externalStringValue,
		project:      anchor.Project,
		groupKey:     anchor.GroupKey,
		paramKey:     anchor.ParamKey,
		valueLabel:   anchor.ValueLabel,
		currentValue: anchor.CurrentValue,
		source:       m.currentValueEditSource(),
	})
}

func (m *Model) openExternalValueEditor(session externalValueEditSession) tea.Cmd {
	stagedValue := session.currentValue
	pattern := "fbrcm-value-*.txt"
	if session.kind == externalJSONValue {
		stagedValue = jsoninput.PrettyJSON(stagedValue)
		pattern = "fbrcm-value-*.json"
	}
	temp, err := os.CreateTemp("", pattern)
	if err != nil {
		m.openErrorDialog("External Editor Failed", session.project, fmt.Sprintf("Create staged value: %v", err))
		return nil
	}
	path := temp.Name()
	if err := temp.Chmod(coreconfig.PrivateFileMode); err != nil {
		_ = temp.Close()
		_ = os.Remove(path)
		m.openErrorDialog("External Editor Failed", session.project, fmt.Sprintf("Secure staged value: %v", err))
		return nil
	}
	if _, err := temp.WriteString(stagedValue); err != nil {
		_ = temp.Close()
		_ = os.Remove(path)
		m.openErrorDialog("External Editor Failed", session.project, fmt.Sprintf("Write staged value: %v", err))
		return nil
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(path)
		m.openErrorDialog("External Editor Failed", session.project, fmt.Sprintf("Close staged value: %v", err))
		return nil
	}

	m.closeOverlays()
	session.path = path
	m.externalEdit = &session
	return m.runExternalValueEditor()
}

func (m Model) runExternalValueEditor() tea.Cmd {
	if m.externalEdit == nil || m.externalEdit.path == "" {
		return nil
	}
	path := m.externalEdit.path
	process := coreeditor.Command(coreeditor.Resolve(""), path)
	return tea.ExecProcess(process, func(err error) tea.Msg {
		return externalValueEditorFinishedMsg{path: path, err: err}
	})
}

func (m Model) updateExternalValueEditorFinished(msg externalValueEditorFinishedMsg) (Model, tea.Cmd, bool) {
	session := m.externalEdit
	if session == nil || session.path != msg.path {
		return m, nil, true
	}
	if msg.err != nil {
		m.openExternalValueEditorRecoveryDialog("External Editor Failed", fmt.Sprintf("The editor failed: %v", msg.err))
		return m, nil, true
	}
	raw, err := os.ReadFile(session.path)
	if err != nil {
		m.openExternalValueEditorRecoveryDialog("External Editor Failed", fmt.Sprintf("Read staged value: %v", err))
		return m, nil, true
	}

	nextValue := string(raw)
	if session.kind == externalJSONValue {
		var decoded any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			m.openExternalValueEditorRecoveryDialog("Invalid JSON", fmt.Sprintf("The edited value is invalid JSON: %v", err))
			return m, nil, true
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, raw); err != nil {
			m.openExternalValueEditorRecoveryDialog("Invalid JSON", fmt.Sprintf("The edited value is invalid JSON: %v", err))
			return m, nil, true
		}
		nextValue = compact.String()
	}

	if err := os.Remove(session.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		m.openExternalValueEditorRecoveryDialog("External Editor Failed", fmt.Sprintf("Remove staged value: %v", err))
		return m, nil, true
	}
	m.externalEdit = nil
	next, cmd := m.applyExternalValueEdit(*session, nextValue)
	return next, cmd, true
}

func (m Model) applyExternalValueEdit(session externalValueEditSession, nextValue string) (Model, tea.Cmd) {
	if nextValue == session.currentValue {
		if session.source == panels.Details {
			m.finishConditionalValueAdd()
		}
		return m, nil
	}
	if session.source == panels.Details {
		m.details = m.details.SetSelectedValue(nextValue)
		m.finishConditionalValueAdd()
		return m, nil
	}
	if session.kind == externalJSONValue {
		if m.parameters.HasDraft(session.project.ProjectID) {
			return m, m.setJSONParameterValueCmd(session.project, session.groupKey, session.paramKey, session.valueLabel, nextValue, false)
		}
		m.openJSONValueDialog(session.project, session.groupKey, session.paramKey, session.valueLabel, nextValue)
		return m, nil
	}
	if m.parameters.HasDraft(session.project.ProjectID) {
		return m, m.setStringParameterValueCmd(session.project, session.groupKey, session.paramKey, session.valueLabel, nextValue, false)
	}
	m.openStringValueDialog(session.project, session.groupKey, session.paramKey, session.valueLabel, nextValue)
	return m, nil
}

func (m *Model) openExternalValueEditorRecoveryDialog(title, problem string) {
	if m.externalEdit == nil {
		return
	}
	path := m.externalEdit.path
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: title,
		Body: []string{
			dialogProjectLine(m.externalEdit.project),
			"",
			problem,
			"",
			"The original value was not changed.",
			"The staged value is preserved at:",
			path,
		},
		Buttons: []dialogcmp.Button{
			{Label: "Reopen", Variant: dialogcmp.ButtonVariantAccent, OnPress: func() tea.Msg { return externalValueEditorReopenMsg{} }},
			{Label: "Close", Variant: dialogcmp.ButtonVariantNeutral, OnPress: func() tea.Msg { return externalValueEditorDismissedMsg{} }},
		},
	})
}

func (m Model) updateExternalValueEditorReopen() (Model, tea.Cmd, bool) {
	return m, m.runExternalValueEditor(), true
}

func (m Model) updateExternalValueEditorDismissed() (Model, tea.Cmd, bool) {
	m.externalEdit = nil
	return m, nil, true
}
