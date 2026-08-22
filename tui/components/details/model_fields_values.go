package details

import (
	"bytes"
	"encoding/json"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/components/inputstyles"
	jsoninput "github.com/yumauri/fbrcm/tui/components/jsoninput"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

const valuePreviewLineBudget = 8

func newTextInput() textinput.Model {
	return inputstyles.NewTextInput()
}

func newDescriptionInput() textarea.Model {
	return inputstyles.NewDescriptionTextarea()
}

func newGroupInput() textinput.Model {
	input := newTextInput()
	input.Placeholder = "New group"
	return input
}

func (m Model) conditionStyle(color string) lipgloss.Style {
	return styles.DetailsConditionValueStyle(color)
}

func (m Model) valueTextStyle(value core.ParametersValue) lipgloss.Style {
	if value.Empty || value.UseInAppDefault {
		return corestyles.EmptyValueStyle()
	}
	return corestyles.ValueTextStyle(value.Value, value.ValueType)
}

func (m Model) renderValueLines(value core.ParametersValue, width int) []string {
	if value.Empty || value.UseInAppDefault {
		return []string{corestyles.EmptyValueStyle().Render(value.Value)}
	}
	if rendered, ok := viewutil.RenderManagedValueSummary(value.Display, value.ValueType); ok {
		return []string{ansi.Truncate(rendered, width, "")}
	}
	switch strings.TrimSpace(strings.ToLower(value.ValueType)) {
	case "json":
		return renderJSONValueLines(value.RawValue, width)
	case "string", "":
		return renderPlainValueLines(value.RawValue, width, corestyles.ValueTextStyle(value.RawValue, value.ValueType))
	}
	return renderPlainValueLines(value.Value, width, m.valueTextStyle(value))
}

func renderPlainValueLines(value string, width int, style lipgloss.Style) []string {
	fragment, cropped := viewutil.CropTextBeforeRender(value, width, width, valuePreviewLineBudget)
	lines := make([]string, 0)
	for part := range strings.SplitSeq(fragment, "\n") {
		for _, line := range wrapLine(part, width) {
			lines = append(lines, style.Render(line))
		}
	}
	if len(lines) == 0 {
		return []string{style.Render("")}
	}
	return finishValuePreview(lines, width, cropped)
}

func renderJSONValueLines(value string, width int) []string {
	fragment, cropped := viewutil.CropTextBeforeRender(value, width, width, valuePreviewLineBudget)
	var out bytes.Buffer
	if cropped {
		out.WriteString(fragment)
	} else if err := json.Indent(&out, []byte(fragment), "", "  "); err != nil {
		return renderPlainValueLines(fragment, width, corestyles.ValueTextStyle(fragment, "json"))
	}
	formatted, formattedCropped := viewutil.CropTextBeforeRender(out.String(), width, width, valuePreviewLineBudget)
	cropped = cropped || formattedCropped
	lines := strings.Split(formatted, "\n")
	ranges := make([]jsoninput.JSONRange, 0, len(lines))
	offset := 0
	for _, line := range lines {
		lineLen := len([]rune(line))
		ranges = append(ranges, jsoninput.JSONRange{Start: offset, End: offset + lineLen, CursorCol: -1})
		offset += lineLen + 1
	}

	highlightedLines := jsoninput.HighlightJSONRanges(formatted, ranges)
	rendered := make([]string, 0, len(highlightedLines))
	for _, highlighted := range highlightedLines {
		indent := min(leadingSpaceWidth(ansi.Strip(highlighted))+2, max(width-1, 0))
		rendered = append(rendered, viewutil.WrapRenderedLine(highlighted, width, indent)...)
	}
	return finishValuePreview(rendered, width, cropped)
}

func finishValuePreview(lines []string, width int, cropped bool) []string {
	if len(lines) > valuePreviewLineBudget {
		lines = append([]string(nil), lines[:valuePreviewLineBudget]...)
		cropped = true
	}
	if !cropped || len(lines) == 0 {
		return lines
	}
	ellipsis := styles.PanelMuted.Render("…")
	lines[len(lines)-1] = ansi.Truncate(lines[len(lines)-1], max(width-lipgloss.Width(ellipsis), 0), "") + ellipsis
	return lines
}
