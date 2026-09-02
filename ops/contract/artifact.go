package contract

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"unicode/utf8"
)

type ArtifactData struct {
	Target      *string         `json:"target"`
	MediaType   string          `json:"media_type"`
	Encoding    string          `json:"encoding" contract:"ref=artifact_encoding"`
	JSONContent json.RawMessage `json:"json_content"`
	TextContent *string         `json:"text_content"`
	Base64      *string         `json:"base64_content"`
	Destination *string         `json:"destination"`
	SizeBytes   int64           `json:"size_bytes"`
	SHA256      string          `json:"sha256"`
	Overwritten bool            `json:"overwritten"`
}

func NewArtifact(target *string, mediaType string, body []byte, destination *string, overwritten bool) ArtifactData {
	result := ArtifactData{Target: target, MediaType: mediaType, Encoding: "base64", Destination: destination, Overwritten: overwritten}
	if destination != nil {
		result.Encoding = "none"
		setArtifactDigest(&result, body)
		return result
	}
	if json.Valid(body) {
		result.Encoding = "json"
		// Marshal RawMessage once so the digest describes the contract's stable
		// form: compact, order-preserving JSON with HTML escaping enabled.
		canonical, err := json.Marshal(json.RawMessage(body))
		if err != nil {
			panic("json.Valid input failed to marshal: " + err.Error())
		}
		result.JSONContent = canonical
		setArtifactDigest(&result, canonical)
		return result
	}
	if utf8.Valid(body) {
		text := string(body)
		result.Encoding = "utf-8"
		result.TextContent = &text
		setArtifactDigest(&result, body)
		return result
	}
	encoded := base64.StdEncoding.EncodeToString(body)
	result.Base64 = &encoded
	setArtifactDigest(&result, body)
	return result
}

func setArtifactDigest(result *ArtifactData, body []byte) {
	sum := sha256.Sum256(body)
	result.SizeBytes = int64(len(body))
	result.SHA256 = hex.EncodeToString(sum[:])
}
