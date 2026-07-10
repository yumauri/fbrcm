package diff

import (
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rootgroup"
)

func formatGroupValue(group string) string {
	if group == "" {
		return rootgroup.Label
	}
	return "[" + group + "]"
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(empty)"
	}
	return value
}

func normalizeJSON(body []byte) []byte {
	return firebase.NormalizeJSONEscapes(body)
}
