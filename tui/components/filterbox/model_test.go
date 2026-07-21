package filterbox

import (
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/tui/styles"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestFilterboxViewActive(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New()
	m.Activate(filter.ModeFuzzy)
	m.input.SetValue("login")

	lines := m.View(40, true, 3)
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
	got := testutil.NormalizeViewSnapshot(strings.Join(lines, "\n"))
	if !strings.Contains(got, "~") || !strings.Contains(got, "login") || !strings.Contains(got, "3") {
		t.Fatalf("view = %q", got)
	}
}

func TestExpressionModeKeepsLastValidExpressionAndUsesErrorColor(t *testing.T) {
	m := New()
	m.ActivateExpression()
	m, _ = m.Update(tea.PasteMsg{Content: `name == "feature_login"`})
	compiled := m.CompiledExpression()
	if compiled == nil || !m.ExpressionValid() {
		t.Fatal("valid expression did not compile")
	}

	m, _ = m.Update(tea.PasteMsg{Content: `name ==`})
	if m.ExpressionValid() {
		t.Fatal("incomplete expression is valid")
	}
	if m.CompiledExpression() != compiled {
		t.Fatal("invalid expression replaced the last valid compiled expression")
	}
	got := m.input.Styles().Focused.Text.GetForeground()
	if want := styles.PaletteError; !reflect.DeepEqual(got, want) {
		t.Fatalf("invalid expression color = %#v, want %#v", got, want)
	}

	m, _ = m.Update(tea.PasteMsg{Content: `name == "login"`})
	got = m.input.Styles().Focused.Text.GetForeground()
	if want := styles.FilterText.GetForeground(); !reflect.DeepEqual(got, want) {
		t.Fatalf("restored valid expression color = %#v, want %#v", got, want)
	}
}

func TestExpressionModeUsesSeparateRememberedBuffer(t *testing.T) {
	m := New()
	m.Activate(filter.ModeIncludes)
	m, _ = m.Update(tea.PasteMsg{Content: "login"})
	m.ActivateExpression()
	if m.Value() != "" {
		t.Fatalf("initial expression value = %q, want empty", m.Value())
	}
	m, _ = m.Update(tea.PasteMsg{Content: `name == "login"`})
	m.Activate(filter.ModeExact)
	if m.Value() != "login" {
		t.Fatalf("restored text value = %q, want login", m.Value())
	}
	m.ActivateExpression()
	if m.Value() != `name == "login"` {
		t.Fatalf("restored expression value = %q", m.Value())
	}
}

func TestInvalidExpressionDiagnosticOverlaysBottomBorderWithoutChangingHeight(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New()
	m.ActivateExpression()
	m, _ = m.Update(tea.PasteMsg{Content: `name ==`})

	if got := m.Height(); got != 2 {
		t.Fatalf("invalid expression height = %d, want 2", got)
	}
	if lines := m.View(60, true, 4); len(lines) != 2 {
		t.Fatalf("invalid expression filter lines = %d, want 2", len(lines))
	}
	base := "╭" + strings.Repeat("─", 58) + "╮\n" +
		"│" + strings.Repeat(" ", 58) + "│\n" +
		"╰" + strings.Repeat("─", 58) + "╯"
	overlaid := m.OverlayExpressionError(base, 1)
	lines := strings.Split(testutil.NormalizeViewSnapshot(overlaid), "\n")
	if len(lines) != 3 {
		t.Fatalf("overlaid panel height = %d, want 3", len(lines))
	}
	bottom := lines[len(lines)-1]
	if !strings.HasPrefix(bottom, "╰──") || !strings.HasSuffix(bottom, "╯") {
		t.Fatalf("overlay did not preserve bottom-border edges: %q", bottom)
	}
	if !strings.Contains(bottom, "Expression error:") || !strings.Contains(bottom, "unexpected token") {
		t.Fatalf("bottom-border overlay = %q", bottom)
	}
	if strings.Contains(bottom, "name ==") {
		t.Fatalf("diagnostic contains multiline source excerpt: %q", bottom)
	}
}

func TestFilterboxClearAndBlur(t *testing.T) {
	m := New()
	m.Activate(filter.ModeIncludes)
	m.input.SetValue("x")
	m.ClearAndBlur()
	if m.Visible() {
		t.Fatal("filter should be hidden after clear")
	}
}

func TestFilterboxPasteSetsValue(t *testing.T) {
	m := New()
	m.Activate(filter.ModeExact)
	updated, _ := m.Update(tea.PasteMsg{Content: "demo"})
	if updated.Value() != "demo" {
		t.Fatalf("value = %q, want demo", updated.Value())
	}
}
