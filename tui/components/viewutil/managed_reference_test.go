package viewutil

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestRenderManagedReferenceIdentityUsesSharedStyles(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	reference := core.ManagedValueReference{
		Group:          "checkout",
		Parameter:      "enabled",
		Condition:      "staff",
		ConditionColor: "PURPLE",
		ValueType:      "BOOLEAN",
	}

	got := RenderManagedReferenceIdentity(reference)
	for field, want := range map[string]string{
		"group":     styles.ParameterGroup.Render("checkout"),
		"parameter": styles.ParameterName.Render("enabled"),
		"condition": styles.DetailsConditionValueStyle("PURPLE").Render("staff"),
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("managed reference does not use the %s style: %q", field, got)
		}
	}
}

func TestRenderManagedReferenceIdentityRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	reference := core.ManagedValueReference{
		Group:          " checkout ",
		Parameter:      " enabled ",
		Condition:      " staff ",
		ConditionColor: "PURPLE",
		ValueType:      " BOOLEAN ",
	}

	if got, want := RenderManagedReferenceIdentity(reference), "checkout / enabled · staff · BOOLEAN"; got != want {
		t.Fatalf("no-color managed reference = %q, want %q", got, want)
	}
}

func TestRenderManagedReferenceConditionStylesDefaultValueSeparately(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := RenderManagedReferenceCondition(core.ManagedValueReference{
		Default: true, Condition: "ignored", ConditionColor: "GREEN",
	})
	want := styles.DetailsEmptyValue.Render("Default value")
	if got != want {
		t.Fatalf("default condition = %q, want %q", got, want)
	}
}
