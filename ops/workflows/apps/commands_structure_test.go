package apps

import (
	"errors"
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestAppsCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "config", "list", "show")
	for _, flag := range []string{"filter", "json", "platform", "show-deleted"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"list"}, flag)
	}
	cmdtest.AssertNestedFlag(t, cmd, []string{"show"}, "json")
	cmdtest.AssertNestedFlag(t, cmd, []string{"config"}, "to")
	cmdtest.AssertNestedFlag(t, cmd, []string{"config"}, "json")
	cmdtest.AssertNestedFlag(t, cmd, []string{"config"}, "yes")
}

func TestParsePlatform(t *testing.T) {
	for input, want := range map[string]core.AppPlatform{"android": core.AppPlatformAndroid, " IOS ": core.AppPlatformIOS, "WEB": core.AppPlatformWeb} {
		got, err := parsePlatform(input)
		if err != nil || got != want {
			t.Errorf("parsePlatform(%q) = %q, %v", input, got, err)
		}
	}
	_, err := parsePlatform("desktop")
	var argumentErr *shared.ArgumentError
	if err == nil || !errors.As(err, &argumentErr) {
		t.Fatalf("invalid platform error = %v", err)
	}
}

func TestFilterAppsMatchesDocumentedFields(t *testing.T) {
	items := []core.FirebaseApp{
		{DisplayName: "Wallet", Namespace: "com.example.wallet", AppID: "android-id"},
		{DisplayName: "Checkout", Namespace: "com.example.pay", AppID: "web-id"},
	}
	for _, filter := range []string{"=Wallet", "=com.example.wallet", "=android-id"} {
		got := filterApps(items, []string{filter})
		if len(got) != 1 || got[0].AppID != "android-id" {
			t.Errorf("filter %q = %#v", filter, got)
		}
	}
}

func TestClassifyAppErrorPreservesTypedSelection(t *testing.T) {
	err := classifyAppError(&core.AppLookupError{Kind: "ambiguous", Query: "Demo", Candidates: []core.AppLookupCandidate{{Name: "Demo", ID: "one"}, {Name: "Demo", ID: "two"}}})
	var selection *shared.SelectionError
	if !errors.As(err, &selection) || selection.Resource != "app" || len(selection.Candidates) != 2 {
		t.Fatalf("selection error = %#v", err)
	}
}
