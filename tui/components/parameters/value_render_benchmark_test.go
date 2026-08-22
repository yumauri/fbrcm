package parameters

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func BenchmarkLargeValueRows(b *testing.B) {
	project := core.Project{ProjectID: "large", Name: "Large"}
	tree := &core.ParametersTree{Groups: []core.ParametersGroup{{
		Key:   "performance",
		Label: "performance",
		Parameters: []core.ParametersEntry{
			{
				Key: "huge_json",
				Values: []core.ParametersValue{{
					Label: "default", Value: `{"items":["` + strings.Repeat("abcdefghij", 56_000) + `"]}`, ValueType: "JSON",
				}},
			},
			{
				Key: "huge_text",
				Values: []core.ParametersValue{{
					Label: "default", Value: strings.Repeat("large multiline value\\n", 10_000), ValueType: "STRING",
				}},
			},
		},
	}}}

	m := New(nil).SetBounds(0, 0, 120, 32).SetActive(true)
	m.projects = []projectState{{project: project, tree: tree}}
	m.projectIndex[project.ProjectID] = 0
	m.groupExpanded[m.groupKey(project.ProjectID, "performance")] = true
	m.syncVisible()

	b.ResetTimer()
	for range b.N {
		_ = m.View(true)
	}
}

func TestLargeValueRowsCropBeforeRendering(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	const tail = "tail must not be rendered"
	project := core.Project{ProjectID: "large", Name: "Large"}
	tree := &core.ParametersTree{Groups: []core.ParametersGroup{{
		Key:   "performance",
		Label: "performance",
		Parameters: []core.ParametersEntry{
			{Key: "huge_json", Values: []core.ParametersValue{{
				Label: "default", Value: `{"payload":"` + strings.Repeat("abcdefghij", 100_000) + tail + `"}`, ValueType: "JSON",
			}}},
			{Key: "huge_text", Values: []core.ParametersValue{{
				Label: "default", Value: strings.Repeat("abcdefghij", 100_000) + tail, ValueType: "STRING",
			}}},
		},
	}}}

	m := New(nil).SetBounds(0, 0, 100, 20).SetActive(true)
	m.projects = []projectState{{project: project, tree: tree}}
	m.projectIndex[project.ProjectID] = 0
	m.groupExpanded[m.groupKey(project.ProjectID, "performance")] = true
	m.syncVisible()

	view := m.View(true)
	if strings.Contains(view, tail) {
		t.Fatalf("large-value view contains cropped tail: %q", view)
	}
	if !strings.Contains(view, "…") {
		t.Fatalf("large-value view has no truncation marker: %q", view)
	}
	if len(view) > 10_000 {
		t.Fatalf("large-value view retained excessive rendered content: %d bytes", len(view))
	}
}
