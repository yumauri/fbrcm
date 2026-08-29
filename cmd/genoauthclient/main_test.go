package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testClientID     = "generated-test-client-id.apps.googleusercontent.com"
	testClientSecret = "generated-test-client-secret"
)

func TestRenderGeneratedSourceMasksPlaintextCredentials(t *testing.T) {
	random := bytes.NewReader(bytes.Repeat([]byte{0xa5}, len(testClientID)+len(testClientSecret)))
	source, err := renderGeneratedSource(random, testClientID, testClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	for _, plaintext := range []string{testClientID, testClientSecret} {
		if bytes.Contains(source, []byte(plaintext)) {
			t.Fatalf("generated source contains plaintext credential %q", plaintext)
		}
	}
	if !bytes.Contains(source, []byte("//go:build fbrcm_google_auth")) {
		t.Fatal("generated source is missing the credentialed-build tag")
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "credentials_generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v", err)
	}
}

func TestMaskValueRoundTripsThroughRuntimeDecoder(t *testing.T) {
	plaintext := []byte(testClientSecret)
	ciphertext, mask, err := maskValue(bytes.NewReader(bytes.Repeat([]byte{0x5a}, len(plaintext))), plaintext)
	if err != nil {
		t.Fatal(err)
	}
	decoded := make([]byte, len(ciphertext))
	for index := range ciphertext {
		decoded[index] = ciphertext[index] ^ mask[index]
	}
	if got := string(decoded); got != string(plaintext) {
		t.Fatalf("decoded value = %q, want %q", got, plaintext)
	}
}

func TestVerifyNoPlaintextDetectsCredentialExposure(t *testing.T) {
	directory := t.TempDir()
	safePath := filepath.Join(directory, "safe-binary")
	if err := os.WriteFile(safePath, []byte("masked application bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyNoPlaintext(directory, testClientID, testClientSecret); err != nil {
		t.Fatalf("safe verification = %v", err)
	}

	exposedPath := filepath.Join(directory, "exposed-binary")
	if err := os.WriteFile(exposedPath, []byte("prefix "+testClientSecret+" suffix"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := verifyNoPlaintext(directory, testClientID, testClientSecret)
	if err == nil || !strings.Contains(err.Error(), "plaintext OAuth client secret") {
		t.Fatalf("exposed verification = %v", err)
	}
}

func TestCredentialsFromEnvironmentRequiresBothValues(t *testing.T) {
	t.Setenv(clientIDEnvironment, testClientID)
	t.Setenv(clientSecretEnvironment, "")
	if _, _, err := credentialsFromEnvironment(); err == nil || !strings.Contains(err.Error(), clientSecretEnvironment) {
		t.Fatalf("missing-secret error = %v", err)
	}

	t.Setenv(clientSecretEnvironment, testClientSecret)
	clientID, clientSecret, err := credentialsFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if clientID != testClientID || clientSecret != testClientSecret {
		t.Fatal("credentialsFromEnvironment returned unexpected values")
	}
}
