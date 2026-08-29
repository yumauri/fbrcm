// Package builtinoauth provides build-scoped Google Desktop OAuth credentials.
package builtinoauth

// decode reconstructs a masked value at runtime. This is intentionally only
// obfuscation: a user who controls the binary can still recover the value.
//
//go:noinline
func decode(ciphertext, mask []byte) string {
	if len(ciphertext) == 0 || len(ciphertext) != len(mask) {
		return ""
	}
	plaintext := make([]byte, len(ciphertext))
	for index := range ciphertext {
		plaintext[index] = ciphertext[index] ^ mask[index]
	}
	return string(plaintext)
}
