package rc

import (
	"bytes"

	"github.com/yumauri/fbrcm/core/firebase"
)

// NormalizeExportJSON makes exported Remote Config JSON stable for diffs and files.
func NormalizeExportJSON(body []byte) []byte {
	body = NormalizeJSONEscapes(body)
	body = reorderConditionalValuesKeysNumericFirst(body)
	return body
}

// NormalizeJSONEscapes restores Go JSON encoder's escaped HTML characters.
func NormalizeJSONEscapes(body []byte) []byte {
	return firebase.NormalizeJSONEscapes(body)
}

// TrimTrailingLineBreaks removes trailing CR/LF bytes from generated output.
func TrimTrailingLineBreaks(body []byte) []byte {
	return bytes.TrimRight(body, "\r\n")
}
