package builtinoauth

import "testing"

func TestDecodeMaskedCredential(t *testing.T) {
	ciphertext := []byte{0x32, 0x15, 0x73, 0x09}
	mask := []byte{0x46, 0x70, 0x00, 0x7d}
	if got := decode(ciphertext, mask); got != "test" {
		t.Fatalf("decode() = %q, want test", got)
	}
}

func TestDecodeRejectsIncompleteMaskedCredential(t *testing.T) {
	for _, test := range []struct {
		name       string
		ciphertext []byte
		mask       []byte
	}{
		{name: "empty"},
		{name: "different lengths", ciphertext: []byte{1}, mask: []byte{1, 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := decode(test.ciphertext, test.mask); got != "" {
				t.Fatalf("decode() = %q, want empty", got)
			}
		})
	}
}
