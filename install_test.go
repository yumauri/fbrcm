package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnixInstallerChecksumVerification(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("POSIX shell unavailable")
	}
	realHash, err := exec.LookPath("sha256sum")
	hashArgs := ""
	if err != nil {
		realHash, err = exec.LookPath("shasum")
		hashArgs = "-a 256"
		if err != nil {
			t.Skip("SHA-256 calculator unavailable")
		}
	}
	installer, err := filepath.Abs("install.sh")
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name      string
		tool      string
		checksum  string
		failedURL string
		wantError string
	}{
		{name: "sha256sum", tool: "sha256sum"},
		{name: "shasum fallback", tool: "shasum"},
		{name: "uppercase binary checksum", tool: "sha256sum", checksum: "binary"},
		{name: "mismatch", tool: "sha256sum", checksum: "mismatch", wantError: "Checksum verification failed"},
		{name: "missing asset", tool: "shasum", checksum: "missing", wantError: "unique valid checksum"},
		{name: "malformed hash", tool: "sha256sum", checksum: "malformed", wantError: "unique valid checksum"},
		{name: "duplicate asset", tool: "sha256sum", checksum: "duplicate", wantError: "unique valid checksum"},
		{name: "archive download fails", tool: "sha256sum", failedURL: "archive", wantError: "download failed"},
		{name: "checksums download fails", tool: "shasum", failedURL: "checksums", wantError: "download failed"},
		{name: "calculator fails", tool: "sha256sum", checksum: "calculator-fails", wantError: "calculator failed"},
		{name: "calculator missing", wantError: "Missing required command: sha256sum or shasum"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			bin := filepath.Join(root, "bin")
			installDir := filepath.Join(root, "install")
			for _, dir := range []string{bin, installDir} {
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			for _, tool := range []string{"awk", "cp", "grep", "install", "mkdir", "rm", "sed", "tr"} {
				path, err := exec.LookPath(tool)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(path, filepath.Join(bin, tool)); err != nil {
					t.Fatal(err)
				}
			}
			writeScript := func(name, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(bin, name), []byte("#!"+shell+"\nset -eu\n"+body+"\n"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			writeScript("uname", `case "$1" in -s) echo Darwin ;; -m) echo arm64 ;; *) exit 1 ;; esac`)
			writeScript("mktemp", `mkdir "$TEST_DOWNLOAD_DIR"; echo "$TEST_DOWNLOAD_DIR"`)
			writeScript("curl", `
out= url=
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) out=$2; shift ;;
    https://*) url=$1 ;;
  esac
  shift
done
case "$url" in
  https://api.github.com/repos/yumauri/fbrcm/releases/latest) echo '{"tag_name": "v1.0.0"}' ;;
  https://github.com/yumauri/fbrcm/releases/download/v1.0.0/checksums.txt)
    if [ "$TEST_FAILED_URL" = checksums ]; then echo 'download failed' >&2; exit 22; fi
    cp "$TEST_CHECKSUMS" "$out" ;;
  https://github.com/yumauri/fbrcm/releases/download/v1.0.0/fbrcm_darwin_arm64.tar.gz)
    if [ "$TEST_FAILED_URL" = archive ]; then echo 'download failed' >&2; exit 22; fi
    cp "$TEST_ARCHIVE" "$out" ;;
  *) echo "unexpected URL: $url" >&2; exit 1 ;;
esac`)
			realTar, err := exec.LookPath("tar")
			if err != nil {
				t.Fatal(err)
			}
			writeScript("tar", `: > "$TEST_EXTRACT_MARKER"; exec "$TEST_REAL_TAR" "$@"`)
			if test.tool != "" {
				body := `if [ "$TEST_CHECKSUM_KIND" = calculator-fails ]; then echo 'calculator failed' >&2; exit 1; fi` + "\n"
				if test.tool == "shasum" {
					body += `[ "$1" = -a ] && [ "$2" = 256 ]; shift 2` + "\n"
				}
				body += `exec "$TEST_REAL_HASH" ` + hashArgs + ` "$@"`
				writeScript(test.tool, body)
			}

			archive := filepath.Join(root, "release.tar.gz")
			writeInstallerArchive(t, archive)
			raw, err := os.ReadFile(archive)
			if err != nil {
				t.Fatal(err)
			}
			checksum := fmt.Sprintf("%x  fbrcm_darwin_arm64.tar.gz\n", sha256.Sum256(raw))
			switch test.checksum {
			case "binary":
				checksum = strings.ToUpper(checksum[:64]) + " *fbrcm_darwin_arm64.tar.gz\n"
			case "mismatch":
				checksum = strings.Repeat("0", 64) + checksum[64:]
			case "missing":
				checksum = strings.ReplaceAll(checksum, "fbrcm_darwin_arm64.tar.gz", "fbrcm_darwin_arm64.tar.gz.other")
			case "malformed":
				checksum = strings.Repeat("z", 64) + checksum[64:]
			case "duplicate":
				checksum += checksum
			}
			checksums := filepath.Join(root, "checksums.txt")
			if err := os.WriteFile(checksums, []byte(checksum), 0o600); err != nil {
				t.Fatal(err)
			}
			downloadDir := filepath.Join(root, "downloads")
			marker := filepath.Join(root, "extracted")
			cmd := exec.Command(shell, installer)
			cmd.Env = append(os.Environ(), "PATH="+bin, "INSTALL_DIR="+installDir,
				"TEST_ARCHIVE="+archive, "TEST_CHECKSUMS="+checksums,
				"TEST_DOWNLOAD_DIR="+downloadDir, "TEST_EXTRACT_MARKER="+marker,
				"TEST_REAL_TAR="+realTar, "TEST_REAL_HASH="+realHash,
				"TEST_FAILED_URL="+test.failedURL, "TEST_CHECKSUM_KIND="+test.checksum)
			output, runErr := cmd.CombinedOutput()
			if test.wantError != "" {
				if runErr == nil || !strings.Contains(string(output), test.wantError) {
					t.Fatalf("error = %v, output = %s; want failure containing %q", runErr, output, test.wantError)
				}
				if _, err := os.Stat(marker); !os.IsNotExist(err) {
					t.Fatalf("archive extraction attempted before verification: %v", err)
				}
				entries, err := os.ReadDir(installDir)
				if err != nil || len(entries) != 0 {
					t.Fatalf("failed verification changed install directory: entries=%v, err=%v", entries, err)
				}
			} else {
				if runErr != nil {
					t.Fatalf("installer failed: %v\n%s", runErr, output)
				}
				for name, want := range map[string]string{"fbrcm": "test binary", "fbrcm-LICENSE.txt": "test license"} {
					got, err := os.ReadFile(filepath.Join(installDir, name))
					if err != nil || string(got) != want {
						t.Fatalf("installed %s = %q, err=%v; want %q", name, got, err, want)
					}
				}
			}
			if _, err := os.Stat(downloadDir); !os.IsNotExist(err) {
				t.Fatalf("temporary downloads remain: %v", err)
			}
		})
	}
}

func writeInstallerArchive(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	for name, body := range map[string]string{"fbrcm": "test binary", "LICENSE": "test license"} {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	for _, closeFile := range []func() error{tw.Close, gz.Close, file.Close} {
		if err := closeFile(); err != nil {
			t.Fatal(err)
		}
	}
}
