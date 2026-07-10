package firebase

import "bytes"

// NormalizeJSONEscapes restores Go JSON encoder's escaped HTML characters.
func NormalizeJSONEscapes(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte(`\u003c`), []byte("<"))
	body = bytes.ReplaceAll(body, []byte(`\u003e`), []byte(">"))
	body = bytes.ReplaceAll(body, []byte(`\u0026`), []byte("&"))
	body = bytes.ReplaceAll(body, []byte(`\u003C`), []byte("<"))
	body = bytes.ReplaceAll(body, []byte(`\u003E`), []byte(">"))
	return body
}
