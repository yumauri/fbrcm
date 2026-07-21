package setup

import (
	"reflect"
	"strings"
	"testing"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestInitialEmptyStateOpensAuthenticationMethods(t *testing.T) {
	m := checkingTestModel(true)

	m, cmd := m.Update(inspectedMsg{state: core.StartupState{Profile: "default"}})

	if cmd != nil {
		t.Fatal("empty startup returned unexpected command")
	}
	if m.mode != modeMethods || !m.mandatory {
		t.Fatalf("mode=%v mandatory=%v, want methods mandatory", m.mode, m.mandatory)
	}
	view := testutil.NormalizeViewSnapshot(m.View(90, 28))
	for _, want := range []string{"Authenticate", "Profile: default", "OAuth desktop login", "Service account", "Existing gcloud credentials"} {
		if !strings.Contains(view, want) {
			t.Fatalf("setup view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Recommended") {
		t.Fatalf("setup view retains static authentication recommendation:\n%s", view)
	}
}

func TestSetupFilePickerMatchesCLIFilePickerPresentation(t *testing.T) {
	m := New(nil)
	if !m.filepicker.ShowPermissions || !m.filepicker.ShowSize {
		t.Fatalf("picker columns: permissions=%v size=%v, want both enabled", m.filepicker.ShowPermissions, m.filepicker.ShowSize)
	}
	if want := filepicker.DefaultStyles(); !reflect.DeepEqual(m.filepicker.Styles, want) {
		t.Fatal("setup picker styles differ from the default styles used by the CLI picker")
	}
}

func TestInitialCachedProjectsOpenWorkspaceWithoutAuth(t *testing.T) {
	m := checkingTestModel(true)
	projects := []core.Project{{Name: "Demo", ProjectID: "demo"}}

	_, cmd := m.Update(inspectedMsg{state: core.StartupState{Profile: "default", Projects: projects}})
	if cmd == nil {
		t.Fatal("cached startup workspace command is nil")
	}
	msg, ok := cmd().(WorkspaceReadyMsg)
	if !ok {
		t.Fatalf("message = %T, want WorkspaceReadyMsg", cmd())
	}
	if msg.Source != "cache" || len(msg.Projects) != 1 || msg.Projects[0].ProjectID != "demo" {
		t.Fatalf("ready = %+v, want cached demo", msg)
	}
	if !msg.CachedOnly {
		t.Fatal("cached startup without auth did not request cached-only notice")
	}
}

func TestInitialAuthWithoutProjectsStartsDiscovery(t *testing.T) {
	m := checkingTestModel(true)
	state := core.StartupState{
		Profile:       "default",
		DefaultAuthID: "main",
		Auth:          []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud}},
	}

	m, cmd := m.Update(inspectedMsg{state: state})

	if m.mode != modeDiscovering || cmd == nil {
		t.Fatalf("mode=%v cmd=%v, want discovering command", m.mode, cmd)
	}
	if m.syncAuthID != "" {
		t.Fatalf("syncAuthID = %q, want all identities", m.syncAuthID)
	}
}

func TestManualSetupShowsConfiguredAccountsAndGeneratedNewID(t *testing.T) {
	m := checkingTestModel(false)
	state := core.StartupState{
		Profile:       "work",
		Profiles:      []string{"default", "work"},
		DefaultAuthID: "default",
		Auth:          []config.AuthEntry{{ID: "default", Type: config.AuthTypeOAuth}},
	}
	m, _ = m.Update(inspectedMsg{state: state})
	if m.mode != modeAccounts || m.mandatory {
		t.Fatalf("mode=%v mandatory=%v, want optional accounts", m.mode, m.mandatory)
	}

	m.cursor = len(m.auth)
	m, _ = m.Update(key(tea.KeyEnter))
	if m.mode != modeMethods {
		t.Fatalf("mode = %v, want methods", m.mode)
	}
	m, _ = m.Update(key(tea.KeyEnter))
	if m.mode != modeIdentity {
		t.Fatalf("mode = %v, want identity", m.mode)
	}
	if got := m.identity.Value(); got != "account-2" {
		t.Fatalf("suggested identity = %q, want account-2", got)
	}
}

func TestProfilesCanCreateAndSwitchWithoutCLI(t *testing.T) {
	m := checkingTestModel(false)
	state := core.StartupState{
		Profile:  "default",
		Profiles: []string{"default", "work"},
		Auth:     []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud}},
	}
	m, _ = m.Update(inspectedMsg{state: state})
	m, _ = m.Update(keyText("ctrl+p"))
	if m.mode != modeProfiles || m.cursor != 0 {
		t.Fatalf("mode=%v cursor=%d, want profiles on default", m.mode, m.cursor)
	}

	m, _ = m.Update(key(tea.KeyDown))
	m, _ = m.Update(key(tea.KeyDown))
	if m.mode != modeProfiles || !m.profileInputSelected() || !m.profileIn.Focused() {
		t.Fatalf("mode=%v cursor=%d focused=%v, want focused inline profile row", m.mode, m.cursor, m.profileIn.Focused())
	}
	m.profileIn.SetValue("personal")
	m, cmd := m.Update(key(tea.KeyEnter))
	if m.mode != modeSwitching || m.profileTo != "personal" || cmd == nil {
		t.Fatalf("mode=%v profileTo=%q cmd=%v, want switching personal", m.mode, m.profileTo, cmd)
	}

	m, cmd = m.Update(profileSwitchedMsg{})
	if m.mode != modeChecking || !m.profileNew || cmd == nil {
		t.Fatalf("mode=%v profileNew=%v cmd=%v, want checking reset", m.mode, m.profileNew, cmd)
	}
}

func TestProfileOverridePinsInteractiveProfileSelection(t *testing.T) {
	m := checkingTestModel(false)
	state := core.StartupState{
		Profile:         "default",
		ProfileOverride: "default",
		Profiles:        []string{"default", "work"},
		Auth:            []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud}},
	}
	m, _ = m.Update(inspectedMsg{state: state})
	m, _ = m.Update(keyText("ctrl+p"))
	if m.mode != modeProfiles || m.profileOverride != "default" {
		t.Fatalf("mode=%v override=%q, want pinned profiles", m.mode, m.profileOverride)
	}
	view := ansi.Strip(m.View(90, 28))
	if !strings.Contains(view, "pinned by FBRCM_PROFILE") || strings.Contains(view, "new profile") {
		t.Fatalf("pinned profile view is misleading:\n%s", view)
	}

	previousCursor := m.cursor
	m, cmd := m.Update(key(tea.KeyDown))
	if cmd != nil || m.cursor != previousCursor || m.mode != modeProfiles {
		t.Fatalf("pinned selection moved: cursor=%d mode=%v cmd=%v", m.cursor, m.mode, cmd)
	}
	m, cmd = m.Update(key(tea.KeyEnter))
	if cmd != nil || m.mode != modeProfiles {
		t.Fatalf("pinned enter = mode:%v cmd:%v, want unchanged profiles", m.mode, cmd)
	}
}

func TestSetupQEmitsGuardedQuitRequest(t *testing.T) {
	m := checkingTestModel(false)
	m.mode = modeAccounts
	_, cmd := m.Update(keyText("q"))
	if cmd == nil {
		t.Fatal("setup q returned nil command")
	}
	if _, ok := cmd().(QuitRequestedMsg); !ok {
		t.Fatalf("setup q message = %T, want QuitRequestedMsg", cmd())
	}
}

func TestNoProjectsOffersRecoveryAndEmptyWorkspace(t *testing.T) {
	m := checkingTestModel(true)
	m.mode = modeDiscovering
	m, _ = m.Update(projectsSyncedMsg{source: "firebase"})
	if m.mode != modeNoProjects {
		t.Fatalf("mode = %v, want no projects", m.mode)
	}

	m.cursor = 2
	_, cmd := m.Update(key(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("open empty workspace command is nil")
	}
	msg, ok := cmd().(WorkspaceReadyMsg)
	if !ok || len(msg.Projects) != 0 {
		t.Fatalf("message = %#v, want empty WorkspaceReadyMsg", msg)
	}
}

func TestSetupErrorShowsRecoveryWithoutOverwritingMessage(t *testing.T) {
	m := checkingTestModel(true)
	m, _ = m.Update(inspectedMsg{err: errTest("decode auth config")})
	if m.mode != modeError || m.failure != failureInspect {
		t.Fatalf("mode=%v failure=%v, want inspect error", m.mode, m.failure)
	}
	view := testutil.NormalizeViewSnapshot(m.View(90, 24))
	for _, want := range []string{"Setup problem", "decode auth config", "Existing configuration was not changed"} {
		if !strings.Contains(view, want) {
			t.Fatalf("error view missing %q:\n%s", want, view)
		}
	}
}

func TestSetupHelpUsesSharedTUIHelpStyles(t *testing.T) {
	got := setupHelp(80,
		[2]string{"enter", "continue"},
		[2]string{"esc", "back"},
	)
	for _, want := range []string{
		styles.FilterText.Render("enter"),
		styles.PanelMuted.Render("continue"),
		styles.PanelMuted.Render(" • "),
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("styled setup help missing %q in %q", want, got)
		}
	}
}

func TestSelectedLineUsesSelectionStyleWithoutMarkerOrPadding(t *testing.T) {
	const label = "OAuth desktop login"
	for _, selected := range []bool{false, true} {
		got := selectedLine(label, selected)
		if plain := ansi.Strip(got); plain != label {
			t.Fatalf("selectedLine(selected=%v) = %q, want %q", selected, plain, label)
		}
	}
}

func TestAuthenticateAndProfileListRowsUseTwoColumnInset(t *testing.T) {
	m := checkingTestModel(true)
	m.mode = modeMethods
	methodLines := m.methodsLines(72)
	for _, index := range []int{4, 7, 10} {
		if got := ansi.Strip(methodLines[index]); !strings.HasPrefix(got, "  ") || strings.HasPrefix(got, "   ") {
			t.Fatalf("authentication row %d inset = %q, want exactly two spaces", index, got)
		}
	}

	m.mode = modeProfiles
	m.profiles = []string{"default", "work"}
	m.profile = "default"
	profileLines := m.profilesLines(72)
	for _, index := range []int{2, 3, 4} {
		if got := ansi.Strip(profileLines[index]); !strings.HasPrefix(got, "  ") || strings.HasPrefix(got, "   ") {
			t.Fatalf("profile row %d inset = %q, want exactly two spaces", index, got)
		}
	}
	if view := ansi.Strip(strings.Join(profileLines, "\n")); strings.Contains(view, "+ new profile") || !strings.Contains(view, "  new profile") {
		t.Fatalf("profiles do not render an inline new profile row:\n%s", view)
	}
}

func TestAccountsUseOneContinuousInsetList(t *testing.T) {
	m := checkingTestModel(false)
	m.mode = modeAccounts
	m.auth = []config.AuthEntry{
		{ID: "personal", Type: config.AuthTypeOAuth},
		{ID: "work", Type: config.AuthTypeGCloud},
	}
	m.defaultID = "personal"
	lines := m.accountsLines(72)

	for _, index := range []int{4, 5, 6} {
		if got := ansi.Strip(lines[index]); !strings.HasPrefix(got, "  ") || strings.HasPrefix(got, "   ") {
			t.Fatalf("account row %d inset = %q, want exactly two spaces", index, got)
		}
	}
	if got := ansi.Strip(lines[6]); got != "  + add authentication" {
		t.Fatalf("add authentication row = %q", got)
	}
}

func TestInlineProfileInputAcceptsTextInsteadOfProfileShortcuts(t *testing.T) {
	m := checkingTestModel(false)
	m.mode = modeProfiles
	m.profiles = []string{"default"}
	m.cursor = len(m.profiles)
	_ = m.profileIn.Focus()

	m, cmd := m.Update(keyText("q"))
	if got := m.profileIn.Value(); got != "q" {
		t.Fatalf("inline profile value = %q, want q", got)
	}
	if m.mode != modeProfiles {
		t.Fatalf("typing q left inline profile input: mode=%v cmd=%v", m.mode, cmd)
	}
	if cmd != nil {
		if _, quitting := cmd().(tea.QuitMsg); quitting {
			t.Fatal("typing q in inline profile input requested quit")
		}
	}
}

func TestCredentialFilePickerSeparatesIdentityAndIndentsPicker(t *testing.T) {
	for _, method := range []authMethod{methodOAuth, methodServiceAccount} {
		m := checkingTestModel(true)
		m.method = method
		m.authID = "default"
		lines := m.fileLines(72)

		if got := ansi.Strip(lines[1]); got != "" {
			t.Fatalf("method %v line before identity = %q, want blank", method, got)
		}
		if got := ansi.Strip(lines[2]); got != "Identity: default" {
			t.Fatalf("method %v identity line = %q", method, got)
		}
		if got := ansi.Strip(lines[3]); got != "" {
			t.Fatalf("method %v line after identity = %q, want blank", method, got)
		}
		pickerIndex := 4
		if method == methodOAuth {
			if got := ansi.Strip(lines[5]); got != "" {
				t.Fatalf("OAuth line before file picker = %q, want blank", got)
			}
			pickerIndex = 6
		}
		if got := ansi.Strip(lines[pickerIndex]); got == "" {
			t.Fatalf("method %v file picker rendered an empty row", method)
		}
		if got := ansi.Strip(lines[pickerIndex+1]); got != "" {
			t.Fatalf("method %v line after file picker = %q, want one blank", method, got)
		}
		if got := ansi.Strip(lines[pickerIndex+2]); got == "" {
			t.Fatalf("method %v help after file picker is empty", method)
		}
		if got := lines[pickerIndex]; strings.HasSuffix(got, "\n ") {
			t.Fatalf("method %v file picker retains a trailing empty line: %q", method, got)
		}

	}
	if got, want := viewutil.IndentLines("first\nsecond", 1), " first\n second"; got != want {
		t.Fatalf("indented file picker = %q, want %q", got, want)
	}
}

func TestOAuthClientShortcutUsesHelpKeyStyle(t *testing.T) {
	m := checkingTestModel(true)
	m.method = methodOAuth
	m.authID = "default"
	view := strings.Join(m.fileLines(72), "\n")

	if !strings.Contains(view, cardMutedStyle.Render("Need one? Press ")+styles.FilterText.Render("o")+cardMutedStyle.Render(" to open Google Cloud OAuth clients.")) {
		t.Fatalf("OAuth client shortcut does not use help key styling:\n%s", view)
	}
}

func TestSetupPanelUsesStandardPopupPadding(t *testing.T) {
	view := ansi.Strip(renderSetupPanel("Authenticate", []string{
		"Connect Google credentials.",
		"",
		selectedLine("OAuth desktop login", true),
	}, 40))
	lines := strings.Split(view, "\n")

	if !strings.HasPrefix(lines[0], "╭─ Authenticate ") || !strings.HasSuffix(lines[0], "╮") {
		t.Fatalf("title is not rendered in the top border: %q", lines[0])
	}
	if lines[1] != "│                                           │" {
		t.Fatalf("top popup padding row = %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "│  Connect Google credentials.") {
		t.Fatalf("first content row lacks standard inset padding: %q", lines[2])
	}
	if !strings.HasPrefix(lines[4], "│  OAuth desktop login") {
		t.Fatalf("selection row lacks standard inset padding: %q", lines[4])
	}
}

func TestAccountsAndProfilesTabsRenderInBorder(t *testing.T) {
	rendered := renderSetupTabsPanel(true, true, []string{"body"}, 48)
	for _, key := range []string{"ˆᵃ", "ˆᵖ"} {
		if !strings.Contains(rendered, key) {
			t.Fatalf("default tab key %q does not use key-hint color", key)
		}
	}
	view := ansi.Strip(rendered)
	first, _, _ := strings.Cut(view, "\n")
	if !strings.HasPrefix(first, "╭─ ˆᵃAccounts ── ˆᵖProfiles ") || !strings.HasSuffix(first, "╮") {
		t.Fatalf("tabs are not rendered in the top border: %q", first)
	}
	selection := styles.TitleStyle(true)
	selectedKey := styles.FilterText.Background(selection.GetBackground()).Render("ˆᵃ")
	if !strings.Contains(rendered, selectedKey) {
		t.Fatal("active Accounts tab does not use selected background with key color")
	}
	unfocused := renderSetupTabsPanel(true, false, []string{"body"}, 48)
	if unfocused == rendered {
		t.Fatal("Accounts tab remains selected while Actions owns focus")
	}
}

func TestAccountsAndProfilesSwitchWithConfiguredKeysTabAndArrows(t *testing.T) {
	state := core.StartupState{
		Profile:  "default",
		Profiles: []string{"default", "work"},
		Auth:     []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud}},
	}
	for _, openProfilesKey := range []string{"ctrl+p", "tab", "right", "left"} {
		m := checkingTestModel(false)
		m, _ = m.Update(inspectedMsg{state: state})
		m, _ = m.Update(keyText(openProfilesKey))
		if m.mode != modeProfiles {
			t.Fatalf("Accounts key %q left mode=%v, want Profiles", openProfilesKey, m.mode)
		}
	}
	for _, openAccountsKey := range []string{"ctrl+a", "tab", "right", "left"} {
		m := checkingTestModel(false)
		m, _ = m.Update(inspectedMsg{state: state})
		m, _ = m.Update(keyText("ctrl+p"))
		m, _ = m.Update(keyText(openAccountsKey))
		if m.mode != modeAccounts {
			t.Fatalf("Profiles key %q left mode=%v, want Accounts", openAccountsKey, m.mode)
		}
	}
}

func TestOpenProfilesHonorsRequestedTabAfterInspection(t *testing.T) {
	m := checkingTestModel(false)
	m.requestedMode = modeProfiles
	m, _ = m.Update(inspectedMsg{state: core.StartupState{
		Profile:  "work",
		Profiles: []string{"default", "work"},
	}})
	if m.mode != modeProfiles || m.cursor != 1 {
		t.Fatalf("mode=%v cursor=%d, want Profiles focused on work", m.mode, m.cursor)
	}
}

func TestAccountDeleteWarnsWhenProjectsAreBound(t *testing.T) {
	m := checkingTestModel(false)
	m.mode = modeAccounts
	m.auth = []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud}}
	m.projects = []core.Project{{ProjectID: "one", AuthID: "main"}, {ProjectID: "two", AuthID: "main"}}

	m, cmd := m.Update(keyText("x"))
	if m.mode != modeAccounts || cmd == nil {
		t.Fatalf("mode=%v cmd=%v, want Accounts plus delete request", m.mode, cmd != nil)
	}
	request, ok := cmd().(AuthDeleteRequestedMsg)
	if !ok || request.AuthID != "main" || request.BoundProjects != 2 {
		t.Fatalf("delete request = %#v, want main with two bound projects", request)
	}
}

func TestProfilesOfferRenameAndProtectActiveProfileFromDelete(t *testing.T) {
	m := checkingTestModel(false)
	m.mode = modeProfiles
	m.profile = "default"
	m.profiles = []string{"default", "work"}
	m.cursor = 1

	m, cmd := m.Update(keyText("r"))
	request, ok := cmd().(ProfileRenameRequestedMsg)
	if m.mode != modeProfiles || !ok || request.Profile != "work" {
		t.Fatalf("rename request = mode:%v request:%#v", m.mode, request)
	}
	m.cursor = 0
	m, cmd = m.Update(keyText("x"))
	errorRequest, ok := cmd().(ErrorRequestedMsg)
	if m.mode != modeProfiles || !ok || !strings.Contains(strings.Join(errorRequest.Body, " "), "is active") {
		t.Fatalf("active delete = mode:%v request:%#v", m.mode, errorRequest)
	}
}

func TestConfirmedProfileDeleteKeepsManagementInPopup(t *testing.T) {
	m := checkingTestModel(false)
	m.mode = modeProfiles

	m, cmd := m.Update(ProfileDeleteConfirmedMsg{Profile: "old"})
	if cmd == nil || m.mode != modeDeletingProfile || m.profileFrom != "old" || m.mandatory {
		t.Fatalf("delete confirmation = cmd:%v mode:%v profile:%q mandatory:%v", cmd != nil, m.mode, m.profileFrom, m.mandatory)
	}
}

func TestOAuthAuthorizationCanBeCanceledBackToFilePicker(t *testing.T) {
	m := checkingTestModel(true)
	m.mode = modeAuthenticating
	m.loginBack = modeFile
	m.loginID = 7
	canceled := false
	m.loginStop = func() { canceled = true }

	m, cmd := m.Update(key(tea.KeyEscape))
	if !canceled {
		t.Fatal("OAuth cancel function was not called")
	}
	if m.mode != modeFile || cmd == nil {
		t.Fatalf("mode=%v cmd=%v, want file picker with clear-screen command", m.mode, cmd)
	}
	if m.loginStop != nil || m.loginID == 7 {
		t.Fatalf("loginStop=%v loginID=%d, want canceled attempt invalidated", m.loginStop != nil, m.loginID)
	}

	m, _ = m.Update(authReadyMsg{loginID: 7, err: errTest("late cancellation")})
	if m.mode != modeFile || m.error != nil {
		t.Fatalf("late auth result changed canceled flow: mode=%v err=%v", m.mode, m.error)
	}
}

func TestOAuthAuthorizationViewShowsCancelHint(t *testing.T) {
	m := checkingTestModel(true)
	m.mode = modeAuthenticating
	m.method = methodOAuth
	m.authID = "default"
	m.auth = []config.AuthEntry{{ID: "default", Type: config.AuthTypeOAuth}}

	view := m.View(90, 24)
	for _, want := range []string{
		"If authorization cannot continue",
		styles.FilterText.Render("esc"),
		styles.PanelMuted.Render("cancel"),
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("OAuth authorization view missing %q", want)
		}
	}
}

func TestProjectDiscoveryCanBeCanceled(t *testing.T) {
	m := checkingTestModel(true)
	m.mode = modeDiscovering
	m.syncBack = modeAccounts
	m.syncID = 11
	canceled := false
	m.syncStop = func() { canceled = true }

	m, cmd := m.Update(key(tea.KeyEscape))
	if !canceled || m.mode != modeAccounts || cmd == nil {
		t.Fatalf("canceled=%v mode=%v cmd=%v, want canceled accounts recovery", canceled, m.mode, cmd)
	}
	m, _ = m.Update(projectsSyncedMsg{syncID: 11, err: errTest("late cancellation")})
	if m.mode != modeAccounts || m.error != nil {
		t.Fatalf("late discovery result changed canceled flow: mode=%v err=%v", m.mode, m.error)
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }

func checkingTestModel(initial bool) Model {
	m := New(nil)
	m.mode = modeChecking
	m.initial = initial
	m.mandatory = initial
	return m
}

func key(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func keyText(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text})
}
