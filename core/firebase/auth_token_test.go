package firebase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestReadCachedTokenMissingEmptyAndCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	tok, err := readCachedToken(path)
	if err != nil || tok != nil {
		t.Fatalf("missing token = %v, %v; want nil, nil", tok, err)
	}

	if err := os.WriteFile(path, []byte("   "), 0o600); err != nil {
		t.Fatalf("write empty token: %v", err)
	}
	tok, err = readCachedToken(path)
	if err != nil || tok != nil {
		t.Fatalf("empty token = %v, %v; want nil, nil", tok, err)
	}

	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write corrupt token: %v", err)
	}
	tok, err = readCachedToken(path)
	if err != nil || tok != nil {
		t.Fatalf("corrupt token = %v, %v; want nil, nil", tok, err)
	}
}

func TestWriteAndReadCachedTokenRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	want := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().UTC().Add(time.Hour),
	}
	if err := writeCachedToken(path, want); err != nil {
		t.Fatalf("writeCachedToken returned error: %v", err)
	}

	got, err := readCachedToken(path)
	if err != nil {
		t.Fatalf("readCachedToken returned error: %v", err)
	}
	if !tokensEqual(want, got) {
		t.Fatalf("read token = %+v, want %+v", got, want)
	}
}

func TestReadCachedTokenDuringWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	initial := &oauth2.Token{
		AccessToken:  "old",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().UTC().Add(time.Hour),
	}
	if err := writeCachedToken(path, initial); err != nil {
		t.Fatalf("writeCachedToken initial returned error: %v", err)
	}

	updated := &oauth2.Token{
		AccessToken:  "new",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().UTC().Add(2 * time.Hour),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 200 {
			if _, err := readCachedToken(path); err != nil {
				t.Errorf("readCachedToken during write returned error: %v", err)
				return
			}
		}
	}()

	for range 50 {
		if err := writeCachedToken(path, updated); err != nil {
			t.Fatalf("writeCachedToken updated returned error: %v", err)
		}
	}
	<-done

	got, err := readCachedToken(path)
	if err != nil {
		t.Fatalf("readCachedToken final returned error: %v", err)
	}
	if !tokensEqual(updated, got) {
		t.Fatalf("final token = %+v, want %+v", got, updated)
	}
}
