package tui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type noColorTestModel struct{}

type noColorQuitMsg struct{}

func (noColorTestModel) Init() tea.Cmd {
	return tea.Tick(10*time.Millisecond, func(time.Time) tea.Msg { return noColorQuitMsg{} })
}

func (m noColorTestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(noColorQuitMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}

func (noColorTestModel) View() tea.View {
	content := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff0000")).Render("styled")
	return tea.NewView(content)
}

func TestProgramOptionsForceNoColorForAnyNonEmptyValue(t *testing.T) {
	for _, value := range []string{"0", "false", "custom", " "} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("NO_COLOR", value)
			var out bytes.Buffer
			options := append(programOptions(), tea.WithWindowSize(20, 2), tea.WithInput(nil), tea.WithOutput(&out))
			program := tea.NewProgram(noColorTestModel{}, options...)
			if _, err := program.Run(); err != nil {
				t.Fatal(err)
			}

			got := out.String()
			if strings.Contains(got, "38;2;255;0;0") || strings.Contains(got, "38;5;") {
				t.Fatalf("TUI output contains color: %q", got)
			}
			if !strings.Contains(got, "\x1b[1m") {
				t.Fatalf("TUI output lost non-color bold styling: %q", got)
			}
		})
	}
}
