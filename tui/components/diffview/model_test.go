package diffview

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestDiffViewUsesEditorSizedModalAndFullWidthPropertyRows(t *testing.T) {
	result := compared(t, dictdiff.Input{
		EntityName: "Flag",
		Left: dictdiff.NamedDictionary{Name: "Development", Properties: dictdiff.Dictionary{
			"value": dictdiff.String("left"),
		}},
		Right: dictdiff.NamedDictionary{Name: "Production", Properties: dictdiff.Dictionary{
			"value": dictdiff.String("right"),
		}},
	})
	m := New().Open(80, 24, result)
	view := testutil.NormalizeViewSnapshot(m.View())
	if width := lipgloss.Width(m.View()); width != 76 {
		t.Fatalf("modal width = %d, want editor width 76", width)
	}
	if height := lipgloss.Height(m.View()); height != 20 {
		t.Fatalf("modal height = %d, want editor height 20", height)
	}
	for _, text := range []string{"Diff", "Flag", "Development", "Production", "▾ value", "left", "right"} {
		if !strings.Contains(view, text) {
			t.Fatalf("diff view misses %q:\n%s", text, view)
		}
	}
	lines := strings.Split(view, "\n")
	if len(lines) < 3 ||
		!strings.Contains(lines[0], "Diff") ||
		strings.Contains(lines[0], "Flag") ||
		!strings.HasPrefix(lines[1], "│ Flag") ||
		!strings.HasPrefix(lines[2], "│ Development") ||
		!strings.Contains(lines[2], "│ Production") {
		t.Fatalf("title, entity, and dictionary headers are not on consecutive rows:\n%s", view)
	}
	propertyLine := lineContaining(view, "▾ value")
	if strings.Count(propertyLine, "│") != 2 {
		t.Fatalf("property header contains an internal column divider:\n%s", propertyLine)
	}
	valueLine := lineContaining(view, "left")
	if !strings.HasPrefix(valueLine, "│ left") || !strings.Contains(valueLine, "│ right") {
		t.Fatalf("diff values do not use one-cell outer and center insets:\n%s", valueLine)
	}
	if !strings.Contains(view, "up/k/down/j property") {
		t.Fatalf("diff footer does not advertise both property navigation directions:\n%s", view)
	}
}

func TestDiffViewPropertyHeadersOmitChangeMarkersAndUseNormalTextColor(t *testing.T) {
	const width = 24
	m := New()
	for _, property := range []dictdiff.Property{
		{Name: "changed", Kind: dictdiff.ChangeChanged},
		{Name: "added", Kind: dictdiff.ChangeAdded},
		{Name: "removed", Kind: dictdiff.ChangeRemoved},
	} {
		want := styles.PanelText.Bold(true).Render(
			viewutil.PadRight("▾ "+property.Name, width),
		)
		if got := m.propertyHeader(property, false, width); got != want {
			t.Errorf("%s property header includes change styling or marker:\ngot  %q\nwant %q", property.Kind, got, want)
		}
	}
}

func TestDiffViewStylesGenericEntityKindAndName(t *testing.T) {
	const width = 96
	m := New()
	m.result.EntityName = "Property: WEB (Shopping cart) / terms_url"
	want := viewutil.PadRight(
		styles.PanelMuted.Render("Property:")+
			styles.PanelText.Bold(true).Render(" WEB (Shopping cart) / terms_url"),
		width,
	)
	if got := m.entityHeader(width); got != want {
		t.Fatalf("generic entity heading styling:\ngot  %q\nwant %q", got, want)
	}

	m.result.LeftName = "Current target: Production (prod)"
	m.result.RightName = "Promotion source: Development (dev)"
	leftWidth, rightWidth := columnWidths(width)
	want = viewutil.PadRight(
		styles.PanelMuted.Render("Current target:")+
			styles.PanelText.Bold(true).Render(" Production (prod)"),
		leftWidth,
	) + styles.PanelBorderInactive.Render("│") + " " + viewutil.PadRight(
		styles.PanelMuted.Render("Promotion source:")+
			styles.PanelText.Bold(true).Render(" Development (dev)"),
		rightWidth,
	)
	if got := m.dictionaryHeader(width); got != want {
		t.Fatalf("generic dictionary heading styling:\ngot  %q\nwant %q", got, want)
	}
}

func TestDiffViewRendersAtomicValueChangesInYellow(t *testing.T) {
	result := compared(t, dictdiff.Input{
		Left:  namedDictionaryWithValue("left", dictdiff.Boolean(true)),
		Right: namedDictionaryWithValue("right", dictdiff.Boolean(false)),
	})
	row := result.Properties[0].Chunks[0].Rows[0]
	left := renderSide(row.Left, row.Kind, true, 12)
	right := renderSide(row.Right, row.Kind, false, 12)
	wantLeft := lipgloss.NewStyle().Foreground(styles.PaletteChanged).Render("true")
	wantRight := lipgloss.NewStyle().Foreground(styles.PaletteChanged).Render("false")
	if len(left) != 1 || left[0].text != wantLeft ||
		len(right) != 1 || right[0].text != wantRight {
		t.Fatalf("atomic boolean diff is not rendered as whole yellow values:\nleft=%#v\nright=%#v", left, right)
	}
}

func TestDiffViewWrapsColumnsAndLimitsWrappedContext(t *testing.T) {
	result := compared(t, dictdiff.Input{
		ContextLines: 3,
		Left: dictdiff.NamedDictionary{Name: "left", Properties: dictdiff.Dictionary{
			"value": dictdiff.String(strings.Repeat("context ", 20) + "\nold"),
		}},
		Right: dictdiff.NamedDictionary{Name: "right", Properties: dictdiff.Dictionary{
			"value": dictdiff.String(strings.Repeat("context ", 20) + "\nnew"),
		}},
	})
	m := New().Open(60, 24, result)
	body := strings.Split(testutil.NormalizeViewSnapshot(m.BodyView(m.contentWidth())), "\n")
	header := 0
	for index, line := range body {
		if strings.Contains(line, "▾ value") {
			header = index
			break
		}
	}
	valueRows := body[header+1:]
	if len(valueRows) != wrappedContextLines+1 {
		t.Fatalf("wrapped context and change use %d rows, want %d:\n%s", len(valueRows), wrappedContextLines+1, strings.Join(body, "\n"))
	}
}

func TestDiffViewChunksDistantChangesInsideOneLongLine(t *testing.T) {
	common := strings.Repeat("same ", 40)
	result := compared(t, dictdiff.Input{
		Left: dictdiff.NamedDictionary{Properties: dictdiff.Dictionary{
			"value": dictdiff.String("prefix LEFT-FIRST " + common + " LEFT-SECOND suffix"),
		}},
		Right: dictdiff.NamedDictionary{Properties: dictdiff.Dictionary{
			"value": dictdiff.String("prefix RIGHT-FIRST " + common + " RIGHT-SECOND suffix"),
		}},
	})
	m := New().Open(60, 30, result)
	body := testutil.NormalizeViewSnapshot(m.BodyView(m.contentWidth()))
	if !strings.Contains(body, "⋯") {
		t.Fatalf("distant inline changes were not split into visual chunks:\n%s", body)
	}
	if rows := len(strings.Split(body, "\n")); rows > 2*(wrappedContextLines*2+2)+3 {
		t.Fatalf("long single-line diff rendered %d rows instead of contextual chunks:\n%s", rows, body)
	}
}

func TestDiffViewSelectsAndCollapsesProperties(t *testing.T) {
	result := compared(t, dictdiff.Input{
		Left: dictdiff.NamedDictionary{Properties: dictdiff.Dictionary{
			"first":  dictdiff.String("left first"),
			"second": dictdiff.String("left second"),
		}},
		Right: dictdiff.NamedDictionary{Properties: dictdiff.Dictionary{
			"first":  dictdiff.String("right first"),
			"second": dictdiff.String("right second"),
		}},
	})
	m := New().Open(80, 24, result)
	expandedRows := len(m.bodyRows(m.contentWidth()))
	m, _ = m.Update(key(tea.KeySpace))
	if !m.collapsed["first"] || len(m.bodyRows(m.contentWidth())) >= expandedRows {
		t.Fatalf("first property was not collapsed: %#v", m.collapsed)
	}
	m, _ = m.Update(key(tea.KeyDown))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want second property", m.cursor)
	}
	m, _ = m.Update(key(tea.KeyEnter))
	if !m.collapsed["second"] {
		t.Fatalf("second property was not collapsed: %#v", m.collapsed)
	}
}

func compared(t *testing.T, input dictdiff.Input) dictdiff.Result {
	t.Helper()
	result, err := dictdiff.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func namedDictionaryWithValue(name string, value dictdiff.Value) dictdiff.NamedDictionary {
	return dictdiff.NamedDictionary{
		Name:       name,
		Properties: dictdiff.Dictionary{"value": value},
	}
}

func lineContaining(text, value string) string {
	for line := range strings.SplitSeq(text, "\n") {
		if strings.Contains(line, value) {
			return line
		}
	}
	return ""
}

func key(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}
