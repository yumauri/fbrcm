package parameters

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestBlameRequestsSelectedParameterFromParameterOrValueRow(t *testing.T) {
	for _, kind := range []visibleNodeKind{nodeParameter, nodeValue} {
		t.Run(visibleNodeKindName(kind), func(t *testing.T) {
			m := parityTestModel()
			for index, node := range m.visible {
				if node.kind == kind && node.paramKey == "feature_login" {
					m.cursor = index
					break
				}
			}

			_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'b', Text: "b"}))
			if cmd == nil {
				t.Fatal("b did not request parameter blame")
			}
			message := cmd()
			request, ok := message.(BlameRequestedMsg)
			if !ok {
				t.Fatalf("b message = %T, want BlameRequestedMsg", message)
			}
			if request.Project.ProjectID != "demo-prod" || request.Parameter != "feature_login" {
				t.Fatalf("blame request = %#v", request)
			}
		})
	}
}

func visibleNodeKindName(kind visibleNodeKind) string {
	if kind == nodeValue {
		return "value"
	}
	return "parameter"
}
