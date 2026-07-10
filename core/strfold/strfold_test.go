package strfold

import (
	"reflect"
	"testing"
)

func TestCompareCaseTieBreak(t *testing.T) {
	values := []string{"beta", "Alpha", "alpha"}
	Sort(values)
	want := []string{"Alpha", "alpha", "beta"}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("Sort = %#v, want %#v", values, want)
	}
}

func TestSortedKeys(t *testing.T) {
	got := SortedKeys(map[string]int{
		"beta":  2,
		"alpha": 1,
		"Gamma": 3,
		"Tango": 4,
		"TIGER": 5,
	})
	want := []string{"alpha", "beta", "Gamma", "Tango", "TIGER"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SortedKeys = %#v, want %#v", got, want)
	}

	if got, want := SortedKeys(map[string]int{"beta": 1, "Alpha": 2, "alpha": 3}), []string{"Alpha", "alpha", "beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SortedKeys case tie-break = %#v, want %#v", got, want)
	}
}

func TestCompareProjectsEmptyNameUsesID(t *testing.T) {
	if got := CompareProjects("", "zebra", "Alpha", "alpha"); got <= 0 {
		t.Fatalf("CompareProjects(empty zebra, Alpha alpha) = %d, want > 0", got)
	}
	if got := CompareProjects("", "beta", "", "zebra"); got >= 0 {
		t.Fatalf("CompareProjects(empty beta, empty zebra) = %d, want < 0", got)
	}
}

func TestSortProjectsCaseInsensitiveNames(t *testing.T) {
	type project struct {
		name string
		id   string
	}
	projects := []project{
		{name: "TIGER", id: "tiger"},
		{name: "beta", id: "beta-2"},
		{name: "Tango", id: "tango"},
		{name: "Alpha", id: "alpha"},
		{name: "Beta", id: "beta-1"},
	}

	SortProjects(projects, func(p project) string { return p.name }, func(p project) string { return p.id })

	wantIDs := []string{"alpha", "beta-1", "beta-2", "tango", "tiger"}
	assertProjectOrder(t, projects, wantIDs, func(p project) string { return p.id })
}

func TestSortProjectsEmptyNameUsesProjectID(t *testing.T) {
	type project struct {
		name string
		id   string
	}
	projects := []project{
		{name: "", id: "zebra"},
		{name: "Alpha", id: "alpha"},
		{name: "", id: "beta"},
	}

	SortProjects(projects, func(p project) string { return p.name }, func(p project) string { return p.id })

	wantIDs := []string{"alpha", "beta", "zebra"}
	assertProjectOrder(t, projects, wantIDs, func(p project) string { return p.id })
}

func TestSortProjectsIDTieBreak(t *testing.T) {
	type project struct {
		name string
		id   string
	}
	projects := []project{
		{name: "Same", id: "proj-b"},
		{name: "same", id: "proj-a"},
	}

	SortProjects(projects, func(p project) string { return p.name }, func(p project) string { return p.id })

	wantIDs := []string{"proj-a", "proj-b"}
	assertProjectOrder(t, projects, wantIDs, func(p project) string { return p.id })
}

func assertProjectOrder[T any](t *testing.T, projects []T, wantIDs []string, id func(T) string) {
	t.Helper()
	if len(projects) != len(wantIDs) {
		t.Fatalf("projects length = %d, want %d", len(projects), len(wantIDs))
	}
	for i, want := range wantIDs {
		if id(projects[i]) != want {
			t.Fatalf("project[%d] = %q, want %q; got order %+v", i, id(projects[i]), want, projects)
		}
	}
}
