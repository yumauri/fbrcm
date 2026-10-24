package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/themesource"
	"github.com/yumauri/fbrcm/ops/contract"
)

const repositoryCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const repositoryRoot = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const repositoryTree = "cccccccccccccccccccccccccccccccccccccccc"

type repositoryHTTPFixture struct {
	files     map[string]string
	reads     []string
	status    int
	ambiguous bool
}

func (f *repositoryHTTPFixture) install(t *testing.T) {
	t.Helper()
	useTemporaryContractState(t)
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	for _, variable := range []string{env.GitHubToken, env.GitHubTokenFallback, env.GitLabToken} {
		t.Setenv(variable, "")
	}
	previous := http.DefaultTransport
	http.DefaultTransport = appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-Goog-User-Project") != "" {
			t.Fatal("Google credentials applied to a repository request")
		}
		provider := "github"
		if req.URL.Host == "gitlab.com" {
			provider = "gitlab"
		}
		respond := func(status int, body string) *http.Response {
			return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": []string{"12"}}, Body: io.NopCloser(strings.NewReader(body))}
		}
		asJSON := func(value any) *http.Response {
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			return respond(200, string(raw))
		}
		if f.status != 0 {
			return respond(f.status, "secret-value"), nil
		}
		if strings.Contains(req.URL.Path, "/commits/") {
			if !strings.HasSuffix(req.URL.Path, "/commits/main") && (!f.ambiguous || !strings.HasSuffix(req.URL.Path, "/commits/main/themes")) {
				if provider == "github" {
					return respond(http.StatusUnprocessableEntity, `{"message":"No commit found for SHA"}`), nil
				}
				return respond(404, ""), nil
			}
			if provider == "gitlab" {
				return respond(200, `{"id":"`+repositoryCommit+`"}`), nil
			}
			return respond(200, `{"sha":"`+repositoryCommit+`","commit":{"tree":{"sha":"`+repositoryRoot+`"}}}`), nil
		}
		if strings.Contains(req.URL.Path, "/git/trees/") {
			if strings.HasSuffix(req.URL.Path, repositoryRoot) {
				return respond(200, `{"tree":[{"path":"themes","type":"tree","mode":"040000","sha":"`+repositoryTree+`"}]}`), nil
			}
			type githubEntry struct {
				Path string `json:"path"`
				Type string `json:"type"`
				Mode string `json:"mode"`
			}
			entries := make([]githubEntry, 0, len(f.files))
			for name := range f.files {
				entries = append(entries, githubEntry{Path: name, Type: "blob", Mode: "100644"})
			}
			return asJSON(struct {
				Tree []githubEntry `json:"tree"`
			}{entries}), nil
		}
		if strings.HasSuffix(req.URL.Path, "/repository/tree") {
			if req.URL.Query().Get("ref") != repositoryCommit {
				t.Fatal("mutable directory listing")
			}
			type gitlabEntry struct {
				Name string `json:"name"`
				Path string `json:"path"`
				Type string `json:"type"`
				Mode string `json:"mode"`
			}
			if req.URL.Query().Get("path") == "" {
				return asJSON([]gitlabEntry{{Name: "themes", Path: "themes", Type: "tree", Mode: "040000"}}), nil
			}
			entries := make([]gitlabEntry, 0, len(f.files))
			for name := range f.files {
				entries = append(entries, gitlabEntry{Name: name, Path: "themes/" + name, Type: "blob", Mode: "100644"})
			}
			return asJSON(entries), nil
		}
		if strings.Contains(req.URL.Path, "/contents/") || strings.Contains(req.URL.Path, "/repository/files/") {
			if req.URL.Query().Get("ref") != repositoryCommit {
				t.Fatal("mutable download")
			}
			name := path.Base(strings.TrimSuffix(req.URL.Path, "/raw"))
			body, ok := f.files[name]
			if !ok {
				t.Fatalf("unexpected file: %s", req.URL)
			}
			f.reads = append(f.reads, name)
			return respond(200, body), nil
		}
		t.Fatalf("unexpected request: %s", req.URL)
		return nil, nil
	})
	t.Cleanup(func() { http.DefaultTransport = previous })
	previousOffline := firebase.IsOffline()
	firebase.SetOfflineMode(false)
	t.Cleanup(func() { firebase.SetOfflineMode(previousOffline) })
}
func repositoryBrowserURL(provider, kind, filename string) string {
	prefix := "https://github.com/acme/pack/"
	if provider == "gitlab" {
		prefix = "https://gitlab.com/company/team/pack/-/"
	}
	result := prefix + kind + "/main/themes"
	if filename != "" {
		result += "/" + filename
	}
	return result
}

func TestRepositoryThemeImportRuntimeEnvelopes(t *testing.T) {
	for _, provider := range []string{"github", "gitlab"} {
		for _, batch := range []bool{false, true} {
			t.Run(fmt.Sprint(provider, batch), func(t *testing.T) {
				fixture := &repositoryHTTPFixture{files: map[string]string{"parent.toml": "[colors]\nprimary=\"1\"\n", "child.toml": "inherits=\"parent\"\n[colors]\nerror=\"2\"\n"}}
				fixture.install(t)
				kind, filename := "blob", "parent.toml"
				if batch {
					kind, filename = "tree", ""
				}
				source := repositoryBrowserURL(provider, kind, filename)
				args := []string{"theme", "import", source + "?plain=1#L1", "--json"}
				if !batch {
					args = append(args, "--name", "renamed")
				}
				envelope, raw := executeJSONContract(t, args...)
				if envelope.Outcome != "success" || envelope.ExitCode != 0 {
					t.Fatalf("envelope = %s", raw)
				}
				validateContractDocument(t, envelope.Schema, raw)
				var document struct {
					Data struct {
						RepositorySource themesource.Provenance `json:"repository_source"`
						Source           string                 `json:"source"`
						Imported         int                    `json:"imported_count"`
					}
				}
				if err := json.Unmarshal(raw, &document); err != nil {
					t.Fatal(err)
				}
				if document.Data.Source != source || document.Data.RepositorySource.Commit != repositoryCommit || document.Data.RepositorySource.Provider != provider {
					t.Fatalf("provenance=%s", raw)
				}
				expected := "renamed"
				if batch {
					expected = "child"
					if document.Data.Imported != 2 {
						t.Fatalf("batch=%s", raw)
					}
				}
				if _, err := config.LoadTheme(expected); err != nil {
					t.Fatal(err)
				}
				if batch {
					// Skipped parents are resolved from the installed version and never downloaded.
					fixture.reads = nil
					fixture.files["parent.toml"] = strings.Repeat("invalid", themesource.MaximumFileBytes)
					fixture.files["child.toml"] = "invalid"
					repeated, repeatedRaw := executeJSONContract(t, "theme", "import", source, "--json")
					if repeated.Outcome != "success" || len(fixture.reads) != 0 || len(repeated.Warnings) != 2 {
						t.Fatalf("repeat=%s, downloads=%v", repeatedRaw, fixture.reads)
					}
					validateContractDocument(t, repeated.Schema, repeatedRaw)
					if err := json.Unmarshal(repeatedRaw, &document); err != nil {
						t.Fatal(err)
					}
					if document.Data.Imported != 0 {
						t.Fatalf("no-op=%s", repeatedRaw)
					}
					fixture.files["fresh.toml"] = "inherits=\"parent\"\n"
					fresh, freshRaw := executeJSONContract(t, "theme", "import", source, "--json")
					if fresh.Outcome != "success" || strings.Join(fixture.reads, ",") != "fresh.toml" {
						t.Fatalf("fresh=%s, downloads=%v", freshRaw, fixture.reads)
					}
					validateContractDocument(t, fresh.Schema, freshRaw)
				}
			})
		}
	}
}

func TestRepositoryThemeImportFailuresConform(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		code   string
		exit   int
	}{
		{"invalid credentials", 401, "auth.credentials_invalid", 4}, {"rate limit", 429, "theme.source_rate_limited", 11},
		{"not found or inaccessible", 404, "theme.source_not_found", 6},
		{"unknown GitHub ref", 422, "theme.source_not_found", 6}, {"server failure", 503, "network.unavailable", 11},
		{"redirect", 302, "theme.source_invalid_response", 8},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := &repositoryHTTPFixture{status: test.status}
			fixture.install(t)
			t.Setenv(env.GitHubToken, "secret-value")
			envelope, raw := executeJSONContract(t, "theme", "import", repositoryBrowserURL("github", "tree", ""), "--json")
			if envelope.ExitCode != test.exit || len(envelope.Errors) != 1 || envelope.Errors[0].Code != test.code {
				t.Fatalf("envelope=%s", raw)
			}
			if strings.Contains(string(raw), "secret-value") {
				t.Fatal("token or response body leaked")
			}
			validateContractDocument(t, envelope.Schema, raw)
			assertNoRepositoryThemes(t)
		})
	}
}
func assertNoRepositoryThemes(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(config.GetThemesDirPath()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected themes directory: %v", err)
	}
}

func TestRepositoryThemeBatchValidationDoesNotInstallPartialBatch(t *testing.T) {
	fixture := &repositoryHTTPFixture{files: map[string]string{"a.toml": "[colors]\nprimary=\"1\"\n", "z.toml": "not valid TOML"}}
	fixture.install(t)
	envelope, raw := executeJSONContract(t, "theme", "import", repositoryBrowserURL("github", "tree", ""), "--json")
	if envelope.ExitCode != 3 || envelope.Errors[0].Code != "theme.invalid" {
		t.Fatalf("envelope=%s", raw)
	}
	validateContractDocument(t, envelope.Schema, raw)
	assertNoRepositoryThemes(t)
}

func TestRepositoryThemeAmbiguityAndRefSchema(t *testing.T) {
	fixture := &repositoryHTTPFixture{files: map[string]string{"aurora.toml": "[colors]\nprimary=\"1\"\n"}, ambiguous: true}
	fixture.install(t)
	source := repositoryBrowserURL("github", "tree", "")
	envelope, raw := executeJSONContract(t, "theme", "import", source, "--json")
	if envelope.ExitCode != 2 || envelope.Errors[0].Code != "theme.source_ambiguous" || len(envelope.Errors[0].Remediation) != 2 {
		t.Fatalf("ambiguity=%s", raw)
	}
	validateContractDocument(t, envelope.Schema, raw)
	for _, remediation := range envelope.Errors[0].Remediation {
		if strings.Contains(strings.Join(remediation.Argv, " "), "secret") {
			t.Fatal(remediation)
		}
	}
	envelope, raw = executeJSONContract(t, "theme", "import", source, "--ref", "main", "--json")
	if envelope.ExitCode != 0 {
		t.Fatalf("disambiguated=%s", raw)
	}
	validateContractDocument(t, envelope.Schema, raw)
	schemaID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:theme.import:input"
	for _, test := range []struct {
		source, ref string
		valid       bool
	}{
		{source, "main", true}, {"HTTPS://GITHUB.COM/acme/pack/tree/main/themes", "main", true}, {repositoryBrowserURL("gitlab", "tree", ""), "feature/colors", true},
		{"./themes", "main", false}, {"https://example.com/theme.toml", "main", false},
		{source, "", false}, {source, "bad ref", false}, {source, strings.Repeat("a", 1025), false},
	} {
		validateContractValue(t, schemaID, map[string]any{"arguments": map[string]any{"source": test.source}, "options": map[string]any{"ref": test.ref}, "stdin": nil}, test.valid)
	}
}

func TestRepositoryThemeInputAndOfflineFailures(t *testing.T) {
	for _, test := range []struct {
		name    string
		args    []string
		offline bool
		code    string
	}{
		{"name on directory", []string{repositoryBrowserURL("github", "tree", ""), "--name", "renamed"}, false, "argument.invalid"},
		{"ref mismatch", []string{repositoryBrowserURL("github", "tree", ""), "--ref", "other"}, false, "argument.invalid"},
		{"empty ref", []string{repositoryBrowserURL("github", "tree", ""), "--ref", ""}, false, "argument.invalid"},
		{"ref with stdin", []string{"--ref", "main"}, false, "argument.invalid"},
		{"offline", []string{repositoryBrowserURL("github", "tree", "")}, true, "network.offline"},
		{"empty directory", []string{repositoryBrowserURL("github", "tree", "")}, false, "argument.invalid"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := &repositoryHTTPFixture{files: map[string]string{}}
			fixture.install(t)
			firebase.SetOfflineMode(test.offline)
			args := append([]string{"theme", "import"}, test.args...)
			args = append(args, "--json")
			envelope, raw := executeJSONContract(t, args...)
			if len(envelope.Errors) != 1 || envelope.Errors[0].Code != test.code {
				t.Fatalf("failure=%s", raw)
			}
			validateContractDocument(t, envelope.Schema, raw)
			assertNoRepositoryThemes(t)
		})
	}
}

func TestRepositoryThemeTypedErrorsAndExitCodes(t *testing.T) {
	for _, test := range []struct {
		kind, code string
		exit       int
	}{
		{"invalid", "argument.invalid", 2}, {"not_found", "theme.source_not_found", 6}, {"permission", "theme.source_permission_denied", 5},
		{"rate_limited", "theme.source_rate_limited", 11}, {"limit", "theme.source_limit", 8}, {"timeout", "network.timeout", 9}, {"unavailable", "network.unavailable", 11}, {"request_failed", "theme.source_invalid_response", 8},
	} {
		t.Run(test.kind, func(t *testing.T) {
			useTemporaryContractState(t)
			root := NewRootForContract("test")
			cmd, _, err := root.Find([]string{"theme", "import"})
			if err != nil {
				t.Fatal(err)
			}
			cause := &themesource.Error{Kind: test.kind, Err: fmt.Errorf("same wording for every kind")}
			envelope := contract.BuildEnvelope(cmd, "test", nil, cause)
			if envelope.ExitCode != test.exit || envelope.Errors[0].Code != test.code {
				t.Fatalf("typed=%#v", envelope)
			}
			validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
		})
	}
}

// Keep cancellation classified by the global machine contract, independently of provider transport.
func TestRepositoryThemeCancellationEnvelope(t *testing.T) {
	root := NewRootForContract("test")
	cmd, _, err := root.Find([]string{"theme", "import"})
	if err != nil {
		t.Fatal(err)
	}
	envelope := contract.BuildEnvelope(cmd, "test", nil, context.Canceled)
	if envelope.Errors[0].Code != "command.canceled" {
		t.Fatal(envelope)
	}
	validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
}

func TestRepositoryThemeSingleCollisionDoesNotDownload(t *testing.T) {
	fixture := &repositoryHTTPFixture{files: map[string]string{"aurora.toml": "[colors]\nprimary=\"1\"\n"}}
	fixture.install(t)
	if _, err := config.ImportTheme("aurora", []byte("[colors]\nprimary=\"2\"\n")); err != nil {
		t.Fatal(err)
	}
	envelope, raw := executeJSONContract(t, "theme", "import", repositoryBrowserURL("github", "blob", "aurora.toml"), "--json")
	if envelope.ExitCode != 7 || len(fixture.reads) != 0 {
		t.Fatalf("collision=%s reads=%v", raw, fixture.reads)
	}
	validateContractDocument(t, envelope.Schema, raw)
	installed, err := os.ReadFile(filepath.Join(config.GetThemesDirPath(), "aurora.toml"))
	if err != nil || !strings.Contains(string(installed), `primary="2"`) {
		t.Fatalf("installed=%s, %v", installed, err)
	}
}
