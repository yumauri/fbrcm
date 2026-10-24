package themesource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/env"
)

const testCommit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const testRoot = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const testTree = "cccccccccccccccccccccccccccccccccccccccc"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
func response(status int, data string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}
}
func jsonResponse(value any) *http.Response {
	raw, _ := json.Marshal(value)
	return response(200, string(raw))
}
func requireKind(t *testing.T, err error, kind string) *Error {
	t.Helper()
	var source *Error
	if !errors.As(err, &source) || source.Kind != kind {
		t.Fatalf("error = %v, want source kind %s", err, kind)
	}
	return source
}

func TestParseBrowserURLs(t *testing.T) {
	for _, test := range []struct {
		raw, ref, provider, repository string
		invalid                        bool
	}{
		{raw: "https://github.com/acme/pack/tree/main/themes?plain=1#readme", provider: "github", repository: "acme/pack"},
		{raw: "https://gitlab.com/company/team/pack/-/blob/main/themes/nord.toml", provider: "gitlab", repository: "company/team/pack"},
		{raw: "https://github.com/acme/pack/tree/feature%2Fcolors/themes", ref: "feature/colors", provider: "github", repository: "acme/pack"},
		{raw: "https://gitlab.com/company/pack/-/tree/v1.0/themes/", provider: "gitlab", repository: "company/pack"},
		{raw: "./100%themes"}, {raw: "./themes"}, {raw: "https://example.com/nord.toml"}, {raw: "https://raw.githubusercontent.com/acme/pack/main/nord.toml"},
		{raw: "https://gitlab.com/company/pack/-/raw/main/nord.toml"},
		{raw: "https://github.com/acme/pack", invalid: true},
		{raw: "http://github.com/acme/pack/tree/main/themes", invalid: true},
		{raw: "https://secret@github.com/acme/pack/tree/main/themes", invalid: true},
		{raw: "https://github.com:443/acme/pack/tree/main/themes", invalid: true},
		{raw: "https://github.com/acme%2Fother/pack/tree/main/themes", invalid: true},
		{raw: "https://github.com/acme/pack/tree/main/../themes", invalid: true},
		{raw: "https://github.com/acme/pack/tree/main/%2e%2e%2fthemes", invalid: true},
		{raw: "https://github.com/acme/pack/tree/main/%00.toml", invalid: true},
		{raw: "https://github.com/acme/pack/tree/main//themes", invalid: true},
		{raw: "https://github.com/acme/pack/tree/main/themes", ref: "other", invalid: true},
		{raw: "https://example.com/nord.toml", ref: "main", invalid: true},
		{raw: "", ref: "main", invalid: true},
	} {
		t.Run(test.raw+test.ref, func(t *testing.T) {
			source, err := Parse(test.raw, test.ref)
			if test.invalid {
				_ = requireKind(t, err, "invalid")
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if test.provider == "" {
				if source != nil {
					t.Fatalf("ordinary source parsed: %#v", source)
				}
				return
			}
			if source.Provider != test.provider || source.Repository != test.repository {
				t.Fatalf("source = %#v", source)
			}
			if strings.ContainsAny(source.URL, "?#") {
				t.Fatalf("presentation URL retained: %s", source.URL)
			}
		})
	}
}

// fixture exercises the real parser, ref resolution, metadata and byte fetching.
type fixture struct {
	t        *testing.T
	provider string
	refs     map[string]string
	files    map[string]string
	modes    map[string]string
	requests []*http.Request
	override func(*http.Request) *http.Response
}

func newFixture(t *testing.T, provider string) *fixture {
	t.Helper()
	t.Setenv(env.GitHubToken, "")
	t.Setenv(env.GitHubTokenFallback, "")
	t.Setenv(env.GitLabToken, "")
	return &fixture{t: t, provider: provider, refs: map[string]string{"main": testCommit}, files: map[string]string{"themes/aurora.toml": "[colors]\nprimary = \"1\"\n"}, modes: make(map[string]string)}
}
func (f *fixture) client() *http.Client { return &http.Client{Transport: roundTripFunc(f.serve)} }
func (f *fixture) serve(req *http.Request) (*http.Response, error) {
	f.requests = append(f.requests, req)
	if f.override != nil {
		if result := f.override(req); result != nil {
			return result, nil
		}
	}
	apiPath := req.URL.Path
	marker := "/commits/"
	if _, ref, ok := strings.Cut(apiPath, marker); ok {
		commit, ok := f.refs[ref]
		if !ok {
			if f.provider == "github" {
				return response(http.StatusUnprocessableEntity, `{"message":"No commit found for SHA: `+ref+`","status":"422"}`), nil
			}
			return response(404, "not found"), nil
		}
		if f.provider == "github" {
			return jsonResponse(struct {
				SHA    string `json:"sha"`
				Commit any    `json:"commit"`
			}{commit, struct {
				Tree any `json:"tree"`
			}{struct {
				SHA string `json:"sha"`
			}{testRoot}}}), nil
		}
		return response(200, `{"id":"`+commit+`"}`), nil
	}
	if strings.Contains(apiPath, "/git/trees/") {
		if strings.HasSuffix(apiPath, testRoot) {
			return jsonResponse(struct {
				Tree []entry `json:"tree"`
			}{[]entry{{Path: "themes", Type: "tree", Mode: "040000", SHA: testTree}}}), nil
		}
		if strings.HasSuffix(apiPath, testTree) {
			entries := f.entries()
			for i := range entries {
				entries[i].Path = entries[i].Name
			}
			return jsonResponse(struct {
				Tree []entry `json:"tree"`
			}{entries}), nil
		}
	}
	if strings.HasSuffix(apiPath, "/repository/tree") {
		if req.URL.Query().Get("ref") == "" || req.URL.Query().Get("ref") == "main" {
			f.t.Fatal("mutable GitLab ref")
		}
		if req.URL.Query().Get("path") == "" {
			return jsonResponse([]entry{{Name: "themes", Path: "themes", Type: "tree", Mode: "040000"}}), nil
		}
		return jsonResponse(f.entries()), nil
	}
	var filename string
	if _, tail, ok := strings.Cut(apiPath, "/contents/"); ok {
		filename = tail
	}
	if _, tail, ok := strings.Cut(apiPath, "/repository/files/"); ok {
		filename = strings.TrimSuffix(tail, "/raw")
	}
	if filename != "" {
		if req.URL.Query().Get("ref") == "main" || req.URL.Query().Get("ref") == "" {
			f.t.Fatal("download uses mutable ref")
		}
		if contents, ok := f.files[filename]; ok {
			return response(200, contents), nil
		}
	}
	f.t.Fatalf("unexpected request %s", req.URL)
	return nil, nil
}
func (f *fixture) entries() []entry {
	items := make([]entry, 0, len(f.files))
	for filename := range f.files {
		mode := "100644"
		if custom := f.modes[filename]; custom != "" {
			mode = custom
		}
		name := strings.TrimPrefix(filename, "themes/")
		items = append(items, entry{Name: name, Path: filename, Type: "blob", Mode: mode})
	}
	return items
}
func (f *fixture) parse(t *testing.T, kind, ref, selectedPath, disambiguator string) *Source {
	t.Helper()
	raw := "https://github.com/acme/pack/" + kind + "/" + ref + "/" + selectedPath
	if f.provider == "gitlab" {
		raw = "https://gitlab.com/company/team/pack/-/" + kind + "/" + ref + "/" + selectedPath
	}
	source, err := Parse(raw, disambiguator)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestSourceSegmentLimit(t *testing.T) {
	_, err := Parse("https://github.com/acme/pack/tree/main/"+strings.Repeat("dir/", 65), "")
	_ = requireKind(t, err, "limit")
}

func TestProvidersReadPinnedSnapshots(t *testing.T) {
	for _, provider := range []string{"github", "gitlab"} {
		for _, ref := range []string{"main", "v1.0", "feature/colors", "0253c6268cff029bc545265b83862c3fd270614b"} {
			for _, kind := range []string{"blob", "tree"} {
				t.Run(provider+"/"+ref+"/"+kind, func(t *testing.T) {
					fixture := newFixture(t, provider)
					fixture.refs = map[string]string{ref: testCommit}
					fixture.files["themes/notes.txt"] = "ignored"
					fixture.files["themes/link.toml"] = "ignored"
					fixture.modes["themes/link.toml"] = "120000"
					selectedPath := "themes"
					if kind == "blob" {
						selectedPath += "/aurora.toml"
					}
					source := fixture.parse(t, kind, ref, selectedPath, "")
					repository, err := NewClient(fixture.client(), source).Resolve(context.Background())
					if err != nil {
						t.Fatal(err)
					}
					if repository.Provenance.RequestedRef != ref || repository.Provenance.Commit != testCommit || repository.Provenance.Path != selectedPath {
						t.Fatalf("provenance = %#v", repository.Provenance)
					}
					// Simulate moving the branch after resolution. Reads must remain pinned.
					fixture.refs[ref] = strings.Repeat("d", 40)
					files, err := repository.Files(context.Background())
					if err != nil || len(files) != 1 || files[0].Name != "aurora.toml" {
						t.Fatalf("files = %#v, %v", files, err)
					}
					raw, err := repository.Read(context.Background(), files[0])
					if err != nil || string(raw) != fixture.files["themes/aurora.toml"] {
						t.Fatalf("read = %q, %v", raw, err)
					}
				})
			}
		}
	}
}

func TestAmbiguousURLRequiresExplicitRef(t *testing.T) {
	for _, provider := range []string{"github", "gitlab"} {
		t.Run(provider, func(t *testing.T) {
			fixture := newFixture(t, provider)
			fixture.refs["main/themes"] = testCommit
			source := fixture.parse(t, "tree", "main", "themes", "")
			_, err := NewClient(fixture.client(), source).Resolve(context.Background())
			problem := requireKind(t, err, "ambiguous")
			if len(problem.Candidates) != 2 || problem.SourceURL != source.URL {
				t.Fatalf("ambiguity = %#v", problem)
			}
			source = fixture.parse(t, "tree", "main", "themes", "main")
			repository, err := NewClient(fixture.client(), source).Resolve(context.Background())
			if err != nil || repository.Provenance.Path != "themes" {
				t.Fatalf("resolve = %#v, %v", repository, err)
			}
		})
	}
}

func TestProviderAuthenticationAndPrecedence(t *testing.T) {
	for _, test := range []struct{ provider, primary, fallback, gitlab, want string }{
		{"github", "primary", "fallback", "gitlab", "primary"},
		{"github", "", "fallback", "gitlab", "fallback"},
		{"github", "", "", "gitlab", ""},
		{"gitlab", "github", "fallback", "gitlab", "gitlab"},
		{"gitlab", "github", "fallback", "", ""},
	} {
		t.Run(fmt.Sprint(test), func(t *testing.T) {
			fixture := newFixture(t, test.provider)
			t.Setenv(env.GitHubToken, test.primary)
			t.Setenv(env.GitHubTokenFallback, test.fallback)
			t.Setenv(env.GitLabToken, test.gitlab)
			source := fixture.parse(t, "tree", "main", "themes", "")
			_, err := NewClient(fixture.client(), source).Resolve(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			for _, req := range fixture.requests {
				if test.provider == "github" {
					want := ""
					if test.want != "" {
						want = "Bearer " + test.want
					}
					if req.Header.Get("Authorization") != want || req.Header.Get("PRIVATE-TOKEN") != "" || req.URL.Host != "api.github.com" {
						t.Fatalf("unsafe github request: %s", req.URL)
					}
				} else if req.Header.Get("PRIVATE-TOKEN") != test.want || req.Header.Get("Authorization") != "" || req.URL.Host != "gitlab.com" {
					t.Fatalf("unsafe gitlab request: %s", req.URL)
				}
			}
		})
	}
}

func TestHTTPFailuresAreTypedAndRedacted(t *testing.T) {
	for _, test := range []struct {
		status                 int
		remaining, retry, want string
	}{
		{401, "", "", "auth"}, {403, "", "", "permission"}, {403, "0", "", "rate_limited"}, {403, "", "12", "rate_limited"},
		{429, "", "12", "rate_limited"}, {404, "", "", "not_found"}, {302, "", "", "request_failed"}, {500, "", "", "unavailable"}, {504, "", "", "timeout"}, {400, "", "", "request_failed"},
	} {
		t.Run(fmt.Sprint(test), func(t *testing.T) {
			fixture := newFixture(t, "github")
			t.Setenv(env.GitHubToken, "secret-value")
			fixture.override = func(*http.Request) *http.Response {
				resp := response(test.status, "secret-value")
				resp.Header.Set("X-RateLimit-Remaining", test.remaining)
				resp.Header.Set("Retry-After", test.retry)
				resp.Header.Set("Location", "https://evil.example/secret-value")
				return resp
			}
			source := fixture.parse(t, "tree", "main", "themes", "")
			_, err := NewClient(fixture.client(), source).Resolve(context.Background())
			problem := requireKind(t, err, test.want)
			if strings.Contains(err.Error(), "secret-value") || len(fixture.requests) != 1 && test.status != 404 {
				t.Fatalf("unsafe failure: %v", err)
			}
			if test.retry != "" && problem.RetryAfterMS != 12000 {
				t.Fatalf("retry = %d", problem.RetryAfterMS)
			}
		})
	}
}

func TestCancellationAndTransportErrorsDoNotLeak(t *testing.T) {
	fixture := newFixture(t, "github")
	source := fixture.parse(t, "tree", "main", "themes", "")
	for _, err := range []error{context.Canceled, fmt.Errorf("secret-value"), &timeoutError{}} {
		client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, err })}, source)
		_, got := client.Resolve(context.Background())
		if err == context.Canceled {
			if !errors.Is(got, context.Canceled) {
				t.Fatal(got)
			}
			continue
		}
		if strings.Contains(got.Error(), "secret-value") {
			t.Fatal(got)
		}
		kind := "unavailable"
		if _, ok := err.(*timeoutError); ok {
			kind = "timeout"
		}
		_ = requireKind(t, got, kind)
	}
}

type timeoutError struct{}

func (*timeoutError) Error() string   { return "secret-value" }
func (*timeoutError) Timeout() bool   { return true }
func (*timeoutError) Temporary() bool { return true }

func TestClientBudgetsAndMalformedResponses(t *testing.T) {
	fixture := newFixture(t, "github")
	source := fixture.parse(t, "tree", "main", "themes", "")
	for _, test := range []struct {
		name string
		body string
		init func(*Client)
		raw  bool
		want string
	}{
		{name: "invalid json", body: "secret-value", want: "request_failed"},
		{name: "metadata response limit", body: strings.Repeat("x", maximumMetadataBytes+1), want: "limit"},
		{name: "file response limit", body: strings.Repeat("x", MaximumFileBytes+1), raw: true, want: "limit"},
		{name: "aggregate metadata limit", body: "{}", init: func(c *Client) { c.metadataBytes = maximumTotalMetadataBytes }, want: "limit"},
		{name: "aggregate file limit", body: "theme", raw: true, init: func(c *Client) { c.downloadBytes = MaximumBatchBytes }, want: "limit"},
		{name: "request limit", init: func(c *Client) { c.requests = maximumRequests }, want: "limit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return response(200, test.body), nil })}, source)
			if test.init != nil {
				test.init(client)
			}
			var err error
			if test.name == "invalid json" {
				var target any
				_, err = client.json(context.Background(), client.base(), &target)
			} else {
				_, _, err = client.get(context.Background(), client.base(), test.raw)
			}
			_ = requireKind(t, err, test.want)
		})
	}
	client := NewClient(fixture.client(), source)
	if client.http.Timeout != 20*time.Second {
		t.Fatalf("timeout=%s", client.http.Timeout)
	}
}

func TestGitLabPaginationIsCompleteAndPinned(t *testing.T) {
	fixture := newFixture(t, "gitlab")
	var pages []string
	fixture.override = func(req *http.Request) *http.Response {
		if !strings.HasSuffix(req.URL.Path, "/repository/tree") || req.URL.Query().Get("path") != "themes" {
			return nil
		}
		pages = append(pages, req.URL.Query().Get("page"))
		if req.URL.Query().Get("ref") != testCommit {
			t.Fatal("pagination is not pinned")
		}
		page := req.URL.Query().Get("page")
		name := "one.toml"
		if page == "2" {
			name = "two.toml"
		}
		resp := jsonResponse([]entry{{Name: name, Path: "themes/" + name, Type: "blob", Mode: "100644"}})
		if page == "1" {
			resp.Header.Set("X-Next-Page", "2")
		}
		return resp
	}
	source := fixture.parse(t, "tree", "main", "themes", "")
	repo, err := NewClient(fixture.client(), source).Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	files, err := repo.Files(context.Background())
	if err != nil || len(files) != 2 || strings.Join(pages, ",") != "1,2" {
		t.Fatalf("files=%v pages=%v err=%v", files, pages, err)
	}
}

func TestDirectoryLimitsAndUnsupportedEntries(t *testing.T) {
	for _, test := range []struct{ name, provider, want string }{
		{"truncated", "github", "limit"}, {"invalid entry", "github", "request_failed"},
		{"invalid pagination", "gitlab", "request_failed"}, {"endless pagination", "gitlab", "limit"},
		{"too many files", "github", "limit"}, {"symlink", "github", "not_found"},
		{"empty directory", "gitlab", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newFixture(t, test.provider)
			switch test.name {
			case "symlink":
				fixture.modes["themes/aurora.toml"] = "120000"
			case "too many files":
				for i := range MaximumFiles {
					fixture.files[fmt.Sprintf("themes/extra%d.toml", i)] = ""
				}
			}
			fixture.override = func(req *http.Request) *http.Response {
				if strings.HasSuffix(req.URL.Path, testTree) {
					if test.name == "truncated" {
						return response(200, `{"truncated":true,"tree":[]}`)
					}
					if test.name == "invalid entry" {
						return response(200, `{"tree":[{"path":"../escape.toml"}]}`)
					}
				}
				if strings.HasSuffix(req.URL.Path, "/repository/tree") && req.URL.Query().Get("path") == "themes" {
					if test.name == "empty directory" {
						return response(200, `[]`)
					}
					if test.name == "invalid pagination" || test.name == "endless pagination" {
						resp := jsonResponse([]entry{{Name: "one.toml", Path: "themes/one.toml", Type: "blob", Mode: "100644"}})
						next := "999"
						if test.name == "endless pagination" {
							n, _ := strconv.Atoi(req.URL.Query().Get("page"))
							next = fmt.Sprint(n + 1)
						}
						resp.Header.Set("X-Next-Page", next)
						return resp
					}
				}
				return nil
			}
			kind, selected := "tree", "themes"
			if test.name == "symlink" {
				kind, selected = "blob", "themes/aurora.toml"
			}
			source := fixture.parse(t, kind, "main", selected, "main")
			repo, err := NewClient(fixture.client(), source).Resolve(context.Background())
			if err == nil {
				var files []File
				files, err = repo.Files(context.Background())
				if test.name == "empty directory" && len(files) != 0 {
					t.Fatal(files)
				}
			}
			if test.want != "" {
				_ = requireKind(t, err, test.want)
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReadSizeBoundaryAndPreflight(t *testing.T) {
	fixture := newFixture(t, "github")
	source := fixture.parse(t, "blob", "main", "themes/aurora.toml", "main")
	repo, err := NewClient(fixture.client(), source).Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	files, err := repo.Files(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	file := files[0]
	file.Size = MaximumFileBytes + 1
	before := len(fixture.requests)
	_, err = repo.Read(context.Background(), file)
	_ = requireKind(t, err, "limit")
	if len(fixture.requests) != before {
		t.Fatal("oversized file was downloaded")
	}
	fixture.files[file.Path] = strings.Repeat("x", MaximumFileBytes)
	file.Size = MaximumFileBytes
	raw, err := repo.Read(context.Background(), file)
	if err != nil || len(raw) != MaximumFileBytes {
		t.Fatalf("boundary=%d,%v", len(raw), err)
	}
}

func TestProviderEncodedFilePaths(t *testing.T) {
	for _, provider := range []string{"github", "gitlab"} {
		t.Run(provider, func(t *testing.T) {
			fixture := newFixture(t, provider)
			filename := "nord dark #1.toml"
			fixture.files = map[string]string{"themes/" + filename: "[colors]\nprimary=\"1\"\n"}
			source := fixture.parse(t, "blob", "main", "themes/nord%20dark%20%231.toml", "main")
			repository, err := NewClient(fixture.client(), source).Resolve(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			files, err := repository.Files(context.Background())
			if err != nil || len(files) != 1 {
				t.Fatalf("files=%v,%v", files, err)
			}
			_, err = repository.Read(context.Background(), files[0])
			if err != nil {
				t.Fatal(err)
			}
			requested := fixture.requests[len(fixture.requests)-1].URL
			if requested.RawQuery != "ref="+testCommit || !strings.Contains(requested.EscapedPath(), "nord%20dark%20%231.toml") {
				t.Fatalf("unsafe encoded path=%s", requested)
			}
			if provider == "gitlab" && !strings.Contains(requested.EscapedPath(), "projects/company%2Fteam%2Fpack/repository/files/themes%2F") {
				t.Fatalf("namespace/file encoding=%s", requested)
			}
		})
	}
}

func TestGitHubRepositorySegmentsAreEscaped(t *testing.T) {
	source, err := Parse("https://github.com/acme/pack%23test/tree/main/themes", "main")
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(nil, source)
	if client.base() != "https://api.github.com/repos/acme/pack%23test" {
		t.Fatalf("base=%s", client.base())
	}
}

func TestGitHubFolderURLIgnoresRejectedAlternativeRef(t *testing.T) {
	fixture := newFixture(t, "github")
	source, err := Parse("https://github.com/yumauri/fbrcm/tree/main/themes", "")
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewClient(fixture.client(), source).Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if repository.Provenance.RequestedRef != "main" || repository.Provenance.Path != "themes" {
		t.Fatalf("provenance = %#v", repository.Provenance)
	}
	probedAlternative := false
	for _, req := range fixture.requests {
		if strings.HasSuffix(req.URL.Path, "/commits/main/themes") {
			probedAlternative = true
		}
	}
	if !probedAlternative {
		t.Fatal("did not check the alternative main/themes ref")
	}
	files, err := repository.Files(context.Background())
	if err != nil || len(files) != 1 {
		t.Fatalf("files = %v, %v", files, err)
	}
	if _, err := repository.Read(context.Background(), files[0]); err != nil {
		t.Fatal(err)
	}
}

func TestGitHubUnknownRefKeeps422Status(t *testing.T) {
	fixture := newFixture(t, "github")
	fixture.refs = map[string]string{}
	source := fixture.parse(t, "tree", "main", "themes", "main")
	_, err := NewClient(fixture.client(), source).Resolve(context.Background())
	problem := requireKind(t, err, "not_found")
	if problem.HTTPStatus != http.StatusUnprocessableEntity || problem.Operation != "resolve_ref" {
		t.Fatalf("problem = %#v", problem)
	}
}

func TestGitHub422OutsideRefLookupRemainsFailure(t *testing.T) {
	for _, stage := range []string{"listing", "download"} {
		t.Run(stage, func(t *testing.T) {
			fixture := newFixture(t, "github")
			fixture.override = func(req *http.Request) *http.Response {
				if stage == "listing" && strings.Contains(req.URL.Path, "/git/trees/") || stage == "download" && strings.Contains(req.URL.Path, "/contents/") {
					return response(http.StatusUnprocessableEntity, "secret-value")
				}
				return nil
			}
			source := fixture.parse(t, "tree", "main", "themes", "")
			repository, err := NewClient(fixture.client(), source).Resolve(context.Background())
			if err == nil {
				var files []File
				files, err = repository.Files(context.Background())
				if err == nil {
					_, err = repository.Read(context.Background(), files[0])
				}
			}
			problem := requireKind(t, err, "request_failed")
			if problem.HTTPStatus != http.StatusUnprocessableEntity || strings.Contains(err.Error(), "secret-value") {
				t.Fatalf("problem = %#v", problem)
			}
		})
	}
}
