package filterbox

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/tui/components/inputstyles"
	"github.com/yumauri/fbrcm/tui/styles"
)

type Model struct {
	mode            filter.Mode
	expression      bool
	textValue       string
	expressionValue string
	compiled        *filter.Expression
	expressionErr   error
	input           textinput.Model
}

func New() Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(inputstyles.TextInput())
	input.Blur()

	return Model{
		mode:  filter.ModeFuzzy,
		input: input,
	}
}

func (m Model) Mode() filter.Mode {
	return m.mode
}

func (m Model) ExpressionMode() bool {
	return m.expression
}

func (m Model) ExpressionFocused() bool {
	return m.expression && m.Focused()
}

func (m Model) CompiledExpression() *filter.Expression {
	return m.compiled
}

func (m Model) ExpressionValid() bool {
	return !m.expression || m.expressionErr == nil
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) Focused() bool {
	return m.input.Focused()
}

func (m Model) Visible() bool {
	return m.Focused() || m.Value() != ""
}

func (m Model) Height() int {
	if !m.Visible() {
		return 0
	}
	return 2
}

func (m *Model) Activate(mode filter.Mode) tea.Cmd {
	if m.expression {
		m.expressionValue = m.input.Value()
		m.input.SetValue(m.textValue)
		m.expression = false
	}
	m.mode = mode
	m.applyInputStyle()
	m.input.CursorEnd()
	return m.input.Focus()
}

func (m *Model) ActivateExpression() tea.Cmd {
	if !m.expression {
		m.textValue = m.input.Value()
		m.input.SetValue(m.expressionValue)
		m.expression = true
	}
	m.validateExpression()
	m.input.CursorEnd()
	return m.input.Focus()
}

func (m *Model) Focus() tea.Cmd {
	m.input.CursorEnd()
	return m.input.Focus()
}

func (m *Model) Blur() {
	m.input.Blur()
}

func (m *Model) ClearAndBlur() {
	m.input.SetValue("")
	if m.expression {
		m.expressionValue = ""
		m.compiled = nil
		m.expressionErr = nil
	} else {
		m.textValue = ""
	}
	m.applyInputStyle()
	m.input.Blur()
}

func (m *Model) SetWidth(width int) {
	m.input.SetWidth(max(width-3, 1))
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.PasteMsg:
		m.input.SetValue(msg.Content)
		m.input.CursorEnd()
		m.valueChanged()
		return m, nil
	case tea.ClipboardMsg:
		m.input.SetValue(msg.Content)
		m.input.CursorEnd()
		m.valueChanged()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.valueChanged()
	return m, cmd
}

func (m *Model) valueChanged() {
	if m.expression {
		m.expressionValue = m.input.Value()
		m.validateExpression()
		return
	}
	m.textValue = m.input.Value()
}

func (m *Model) validateExpression() {
	compiled, err := filter.CompileExpression(m.input.Value())
	m.expressionErr = err
	if err == nil {
		m.compiled = compiled
	}
	m.applyInputStyle()
}

func (m *Model) applyInputStyle() {
	inputStyle := inputstyles.TextInput()
	if m.expression && m.expressionErr != nil {
		errorStyle := lipgloss.NewStyle().Foreground(styles.PaletteError)
		inputStyle.Focused.Text = errorStyle
		inputStyle.Blurred.Text = errorStyle
		inputStyle.Cursor.Color = styles.PaletteError
	}
	m.input.SetStyles(inputStyle)
}

func (m Model) View(width int, active bool, count int) []string {
	if !m.Visible() || width <= 0 {
		return nil
	}

	borderStyle := styles.BorderStyle(active)
	textStyle := styles.FilterText
	innerWidth := max(width-1, 0)
	m.SetWidth(innerWidth)

	countText := textStyle.Render(" " + fmt.Sprintf("%d", count) + " ")
	countWidth := lipgloss.Width(countText)
	leftWidth := max(width-countWidth-2, 0)
	rightWidth := max(width-leftWidth-countWidth-1, 0)
	separator := borderStyle.Render(strings.Repeat("─", leftWidth)) +
		countText +
		borderStyle.Render(strings.Repeat("─", rightWidth)) +
		borderStyle.Render("┤")

	label := m.mode.Label()
	if m.expression {
		label = ":"
	}
	content := textStyle.Render(label+" ") + m.input.View()
	content = truncateStyled(content, innerWidth)
	content += strings.Repeat(" ", max(innerWidth-lipgloss.Width(content), 0))
	content += borderStyle.Render("│")

	return []string{separator, content}
}

// OverlayExpressionError composites the expression diagnostic over an already
// rendered panel's bottom border without changing the filter or panel height.
// leftInset is the number of cells before the filter label in the panel.
func (m Model) OverlayExpressionError(panel string, leftInset int) string {
	if panel == "" || !m.expression || m.expressionErr == nil {
		return panel
	}

	width := lipgloss.Width(panel)
	height := lipgloss.Height(panel)
	x := max(leftInset, 0) + 2
	available := max(width-x-1, 0)
	if height == 0 || available == 0 {
		return panel
	}

	message := "Expression error: " + firstLine(m.expressionErr.Error())
	message = ansi.Truncate(message, available, "…")
	message = lipgloss.NewStyle().Foreground(styles.PaletteError).Render(message)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(panel).ID("filter-panel"),
		lipgloss.NewLayer(message).ID("filter-expression-error").X(x).Y(height-1).Z(1),
	).Render()
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(value, "\n")
	return strings.TrimSpace(line)
}

func truncateStyled(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}

	var builder strings.Builder
	current := 0
	for _, r := range value {
		next := current + lipgloss.Width(string(r))
		if next > width {
			break
		}
		builder.WriteRune(r)
		current = next
	}
	return builder.String()
}
