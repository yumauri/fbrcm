package firebase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/yumauri/fbrcm/core/config"
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

func TestWriteCachedTokenIgnoresFixedTemporaryPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	// Another process's staging path must neither block this save nor be removed.
	fixedTemp := path + ".tmp"
	if err := os.Mkdir(fixedTemp, 0o700); err != nil {
		t.Fatal(err)
	}
	want := &oauth2.Token{AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer"}
	if err := writeCachedToken(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := readCachedToken(path)
	if err != nil || !tokensEqual(want, got) {
		t.Fatalf("saved token = %+v, err = %v; want %+v", got, err, want)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != config.PrivateFileMode {
		t.Fatalf("token file permissions: info = %v, err = %v", info, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name() != "token.json" || entries[1].Name() != "token.json.tmp" || !entries[1].IsDir() {
		t.Fatalf("save changed another staging path or left temporary files: %v", entries)
	}
}

func TestWriteCachedTokenConcurrentProcesses(t *testing.T) {
	const pathEnv = "FBRCM_TEST_CONCURRENT_TOKEN_PATH"
	payload := strings.Repeat("token", 16*1024)
	if path := os.Getenv(pathEnv); path != "" {
		// Wait until the parent has launched every writer. Each process has its
		// own tokenCacheMu, so this exercises filesystem coordination directly.
		if _, err := io.ReadFull(os.Stdin, make([]byte, 1)); err != nil {
			t.Fatal(err)
		}
		for i := range 20 {
			prefix := fmt.Sprintf("%d-%d", os.Getpid(), i)
			if err := writeCachedToken(path, &oauth2.Token{
				AccessToken: prefix + payload, RefreshToken: prefix, TokenType: "Bearer",
			}); err != nil {
				t.Fatal(err)
			}
			got, err := readCachedToken(path)
			if err != nil || got == nil || got.AccessToken != got.RefreshToken+payload || got.TokenType != "Bearer" {
				t.Fatalf("concurrent read did not return a complete token: %v", err)
			}
		}
		return
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	type writer struct {
		cmd    *exec.Cmd
		stdin  io.WriteCloser
		output bytes.Buffer
	}
	writers := make([]*writer, 0, 4)
	for range 4 {
		w := &writer{cmd: exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWriteCachedTokenConcurrentProcesses$")}
		w.cmd.Env = append(os.Environ(), pathEnv+"="+path)
		w.cmd.Stdout = &w.output
		w.cmd.Stderr = &w.output
		var err error
		w.stdin, err = w.cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := w.cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if w.cmd.ProcessState == nil {
				_ = w.cmd.Wait()
			}
		})
		writers = append(writers, w)
	}
	for _, w := range writers {
		if _, err := w.stdin.Write([]byte("g")); err != nil {
			t.Error(err)
		}
		if err := w.stdin.Close(); err != nil {
			t.Error(err)
		}
	}
	for _, w := range writers {
		if err := w.cmd.Wait(); err != nil {
			t.Errorf("concurrent writer failed: %v\n%s", err, &w.output)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "token.json" {
		t.Fatalf("unexpected files after concurrent saves: %v", entries)
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
