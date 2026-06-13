package get

import "testing"

func TestCountOutputTotals(t *testing.T) {
	rows := []parameterRow{
		{
			ProjectID: "project-a",
			Key:       "flag_a",
			ValueLines: []valueLine{
				{Value: "default"},
				{Value: "conditional"},
				{Missing: true},
			},
		},
		{
			ProjectID: "project-a",
			Key:       "flag_b",
			ValueLines: []valueLine{
				{Value: "default"},
			},
		},
		{
			ProjectID: "project-b",
			Key:       "   ",
			ValueLines: []valueLine{
				{Missing: true},
			},
		},
	}

	if got := countOutputProjects(rows); got != 2 {
		t.Fatalf("countOutputProjects = %d, want 2", got)
	}
	if got := countOutputParameters(rows); got != 2 {
		t.Fatalf("countOutputParameters = %d, want 2", got)
	}
	if got := countOutputValues(rows); got != 3 {
		t.Fatalf("countOutputValues = %d, want 3", got)
	}
}
