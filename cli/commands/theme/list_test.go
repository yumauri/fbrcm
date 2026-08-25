package theme

import (
	"testing"

	coreconfig "github.com/yumauri/fbrcm/core/config"
)

func TestThemeListAlwaysStartsWithBuiltIn(t *testing.T) {
	items := newThemeListItems([]string{"firebase", "nord"}, "")
	if len(items) != 3 || items[0].Theme != coreconfig.BuiltInThemeName || !items[0].BuiltIn || !items[0].Active {
		t.Fatalf("built-in list item = %#v", items)
	}
	items = newThemeListItems([]string{"firebase"}, "firebase")
	if items[0].Active || !items[1].Active || items[1].BuiltIn {
		t.Fatalf("selected file list items = %#v", items)
	}
}
