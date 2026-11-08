package app

import (
	"fmt"
	"io"
	"maps"
	"os"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/tui/components/contextpanel"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
)

type contextCaptureAction int

const (
	contextChooseCapture contextCaptureAction = iota + 1
	contextCancelCapture
)

type contextCapturePastedMsg string

type contextCaptureLoad struct {
	path string
}

type contextCaptureLoadedMsg struct {
	load      *contextCaptureLoad
	exchanges []evaluate.Exchange
	err       error
}

func (m *Model) openContextImportDialog() {
	body := []string{
		"Paste a Firebase fetch cURL command or HAR content here, or choose a file.",
	}
	if m.requestContext.list.Len() > 0 {
		body = append(body, "The selected request will replace the current inputs.")
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Import Context", Body: body,
		OnPaste: func(content string) tea.Msg { return contextCapturePastedMsg(content) },
		Buttons: []dialogcmp.Button{
			{Label: "Choose File", OnPress: func() tea.Msg { return contextChooseCapture }},
			{Label: "Cancel"},
		},
	})
}

func (m Model) updateContextCaptureAction(action contextCaptureAction) (Model, tea.Cmd, bool) {
	var cmd tea.Cmd
	switch action {
	case contextChooseCapture:
		m.projectIO, cmd = m.projectIO.OpenCapture()
	case contextCancelCapture:
		m.requestContext.reading = nil
	}
	return m, cmd, true
}

func (m Model) beginContextCapture(path, content string) (Model, tea.Cmd, bool) {
	load := &contextCaptureLoad{path: path}
	m.requestContext.reading = load
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Import Context", Body: []string{"Reading capture…"},
		Buttons: []dialogcmp.Button{{Label: "Cancel", OnPress: func() tea.Msg { return contextCancelCapture }}},
	})
	return m, func() tea.Msg {
		var reader io.Reader = strings.NewReader(content)
		if load.path != "" {
			file, err := os.Open(load.path)
			if err != nil {
				return contextCaptureLoadedMsg{load: load, err: err}
			}
			defer func() { _ = file.Close() }()
			reader = file
		}
		exchanges, err := evaluate.ReadCapture(reader)
		return contextCaptureLoadedMsg{load: load, exchanges: exchanges, err: err}
	}, true
}

func (m Model) updateContextCaptureLoaded(msg contextCaptureLoadedMsg) (Model, tea.Cmd, bool) {
	c := &m.requestContext
	if c.reading != msg.load || !m.dialog.IsOpen() {
		return m, nil, true
	}
	c.reading = nil
	m.closeDialog()
	if msg.err != nil || len(msg.exchanges) == 0 {
		body := "The capture contains no Firebase fetch requests."
		if msg.err != nil {
			body = msg.err.Error()
		}
		m.dialog = m.dialog.Open(dialogcmp.Config{
			Title: "Could Not Import Context", Body: []string{body},
			Buttons: []dialogcmp.Button{{Label: "Close"}},
		})
		return m, nil, true
	}
	source := msg.load.path
	if source == "" {
		source = "clipboard"
	}
	if len(msg.exchanges) == 1 {
		m.useContextCapture(msg.exchanges[0], source)
		return m, nil, true
	}
	rows := make([]contextpanel.Row, 0, len(msg.exchanges))
	for i, exchange := range msg.exchanges {
		label := fmt.Sprintf("Entry %d", exchange.Entry)
		detail := exchange.TemplateTarget().String()
		if !exchange.StartedAt.IsZero() {
			detail += " · " + exchange.StartedAt.Format(time.RFC3339)
		}
		if app := exchange.Facts["app.id"]; app.Known && !app.Absent {
			detail += " · " + app.Value
		}
		rows = append(rows, contextpanel.Row{Key: strconv.Itoa(i), Label: label, Value: detail})
	}
	c.exchanges, c.exchangeSource = msg.exchanges, source
	c.pickerBack, c.overlay = c.overlay, contextRequests
	c.picker = contextpanel.Model{}.SetRows(rows, "")
	return m, nil, true
}

func (m *Model) useContextCapture(exchange evaluate.Exchange, source string) {
	c := &m.requestContext
	c.scope = exchange.TemplateTarget()
	c.facts = maps.Clone(exchange.Facts)
	exchange.Facts = nil
	c.capture, c.captureSource = &exchange, source
	c.paused = false
	m.refreshContext(contextScopeKey)
}
