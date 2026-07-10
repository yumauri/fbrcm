package parameters

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func TestFormatRemoteConfigDisplayValue(t *testing.T) {
	value := firebase.RemoteConfigValue{Value: "enabled"}
	if got := FormatRemoteConfigDisplayValue(value, "STRING"); got != rcdisplay.FormatSummary(value, "STRING") {
		t.Fatalf("FormatRemoteConfigDisplayValue() = %q, want %q", got, rcdisplay.FormatSummary(value, "STRING"))
	}
}
