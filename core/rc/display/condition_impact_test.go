package display

import "testing"

func TestConditionImpactFormattingPluralizesCounts(t *testing.T) {
	if got := FormatConditionMoveImpact(1, 2); got != "Priority impact: crosses 1 condition and can change the winning value for 2 parameters." {
		t.Fatalf("move impact = %q", got)
	}
	if got := FormatConditionDeleteImpact(2, 1); got != "Deletion impact: removes 2 conditional values; 1 parameter will have no remaining value and will also be removed." {
		t.Fatalf("delete impact = %q", got)
	}
}
