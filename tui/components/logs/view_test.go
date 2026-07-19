package logs

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestRenderLogsPanelEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := testutil.NormalizeViewSnapshot(renderLogsPanel(nil, 60, 5, true, true, charmlog.InfoLevel, true, false))
	if !strings.Contains(got, "Logs") || !strings.Contains(got, "live") {
		t.Fatalf("panel = %q", got)
	}
}

func TestRenderLogsPanelWithBody(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	body := []string{"line one", "line two"}
	got := testutil.NormalizeViewSnapshot(renderLogsPanel(body, 50, 4, false, false, charmlog.DebugLevel, false, false))
	if !strings.Contains(got, "line one") || !strings.Contains(got, "scroll") {
		t.Fatalf("panel = %q", got)
	}
}

func TestLevelLabel(t *testing.T) {
	if got := levelLabel(charmlog.ErrorLevel); got != "ERROR" {
		t.Fatalf("levelLabel = %q", got)
	}
}

func TestIsBackgroundMessage(t *testing.T) {
	for _, msg := range []tea.Msg{messages.LogLineMsg{Line: "line"}, statusFlashTickMsg{}} {
		if !IsBackgroundMessage(msg) {
			t.Fatalf("IsBackgroundMessage(%T) = false", msg)
		}
	}
	if IsBackgroundMessage(tea.KeyPressMsg(tea.Key{Code: 'j'})) {
		t.Fatal("key press classified as a background logs message")
	}
}
