package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/ops/contract"
)

func TestMultiProjectListsRuntimeContract(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.NoColor, "1")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{
		{Name: "Beta", ProjectID: "beta", AuthID: "main", Templates: []rctarget.Kind{rctarget.Client, rctarget.Server}, PrimaryTemplate: rctarget.Client},
		{Name: "Alpha", ProjectID: "alpha", AuthID: "main"},
	}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"alpha", "beta", "server@beta"} {
		if err := config.SaveParametersCache(id, &config.ParametersCache{CachedAt: time.Now(), RemoteConfig: json.RawMessage(`{"conditions":[{"name":"first","expression":"true"},{"name":"second","expression":"false"}],"parameters":{},"version":{"versionNumber":"7"}}`)}); err != nil {
			t.Fatal(err)
		}
	}
	for _, project := range projects {
		raw, err := json.Marshal([]core.FirebaseApp{{DisplayName: "App", AppID: "1:123:web:abc", Platform: core.AppPlatformWeb, State: "ACTIVE"}})
		if err != nil {
			t.Fatal(err)
		}
		if err := config.SaveAppsIndexCache(project.ProjectID, time.Now(), raw); err != nil {
			t.Fatal(err)
		}
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range []string{"apps", "conditions"} {
		t.Run(family, func(t *testing.T) {
			for _, tc := range []struct {
				name      string
				selectors []string
				count     int
				code      string
			}{
				{"all", nil, map[string]int{"apps": 2, "conditions": 6}[family], ""},
				{"single", []string{"alpha"}, map[string]int{"apps": 1, "conditions": 2}[family], ""},
				{"repeated", []string{"-p", "=alpha", "-p", "=beta", "-p", "=alpha"}, map[string]int{"apps": 2, "conditions": 6}[family], ""},
				{"one-filter", []string{"-p", "=alpha"}, map[string]int{"apps": 1, "conditions": 2}[family], ""},
				{"empty", []string{"-p", "=missing"}, 0, ""},
				{"missing", []string{"missing"}, 0, "project.not_found"},
				{"ambiguous", []string{"/a"}, 0, "project.ambiguous"},
				{"conflict", []string{"alpha", "-p", "=beta"}, 0, "argument.invalid"},
				{"item-filter", []string{"--filter", "=missing"}, 0, ""},
			} {
				t.Run(tc.name, func(t *testing.T) {
					cmd := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
					args := append([]string{family, "list"}, tc.selectors...)
					if family == "apps" {
						args = append(args, "--cached")
					}
					args = append(args, "--json")
					var out bytes.Buffer
					cmd.SetOut(&out)
					cmd.SetErr(&bytes.Buffer{})
					cmd.SetArgs(args)
					cmd.SilenceUsage = true
					cmd.SilenceErrors = true
					executed, err := cmd.ExecuteC()
					if (err != nil) != (tc.code != "") {
						t.Fatalf("error=%v output=%s", err, out.String())
					}
					rawEnvelope, marshalErr := json.Marshal(contract.BuildEnvelope(executed, "test", out.Bytes(), err))
					if marshalErr != nil {
						t.Fatal(marshalErr)
					}
					var envelope map[string]any
					if err := json.Unmarshal(rawEnvelope, &envelope); err != nil {
						t.Fatalf("%v: %s", err, out.String())
					}
					validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:"+family+".list:response", envelope, true)
					if tc.code != "" {
						if !strings.Contains(string(rawEnvelope), `"code":"`+tc.code+`"`) {
							t.Fatalf("output=%s", out.String())
						}
						return
					}
					data := envelope["data"].(map[string]any)
					if data["count"] != float64(tc.count) {
						t.Fatalf("data=%#v", data)
					}
					items := data["items"].([]any)
					for _, raw := range items {
						item := raw.(map[string]any)
						if item["project"] == "" || item["project"] == nil || item["project_id"] == "" || item["project_id"] == nil {
							t.Fatalf("missing identity: %#v", item)
						}
					}
					if len(items) > 0 && items[0].(map[string]any)["project_id"] != "alpha" {
						t.Fatalf("order=%#v", items)
					}
				})
			}
			for _, selectors := range [][]string{nil, {"alpha"}, {"-p", "=alpha"}} {
				cmd := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
				args := append([]string{family, "list"}, selectors...)
				if family == "apps" {
					args = append(args, "--cached")
				}
				var out bytes.Buffer
				cmd.SetOut(&out)
				cmd.SetErr(&bytes.Buffer{})
				cmd.SetArgs(args)
				if err := cmd.Execute(); err != nil {
					t.Fatal(err)
				}
				hasProjectColumn := strings.Contains(out.String(), "│ Project ")
				wantProjectColumn := len(selectors) == 0 || selectors[0] == "-p"
				if hasProjectColumn != wantProjectColumn {
					t.Fatalf("selectors=%v output=%s", selectors, out.String())
				}
			}
		})
	}
	// Missing app data attempts a fetch; unavailable authentication remains a typed failure.
	if err := os.RemoveAll(config.GetAppsCacheDirPath()); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"apps", "list", "--cached", "--json"})
	cmd.SilenceUsage = true
	executed, executeErr := cmd.ExecuteC()
	if executeErr == nil {
		t.Fatal("missing cache succeeded")
	}
	rawEnvelope, err := json.Marshal(contract.BuildEnvelope(executed, "test", out.Bytes(), executeErr))
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(rawEnvelope, &envelope); err != nil {
		t.Fatal(err)
	}
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:apps.list:response", envelope, true)
}

func TestMultiProjectListInputSchemas(t *testing.T) {
	for _, family := range []string{"apps", "conditions"} {
		id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + family + ".list:input"
		for _, stateless := range []bool{false, true} {
			input := func(args, options map[string]any) map[string]any {
				options["stateless"] = stateless
				return map[string]any{"arguments": args, "options": options, "stdin": nil}
			}
			validateContractValue(t, id, input(map[string]any{}, map[string]any{}), true)
			validateContractValue(t, id, input(map[string]any{"project": "alpha"}, map[string]any{}), true)
			validateContractValue(t, id, input(map[string]any{}, map[string]any{"project": []any{"=alpha", "^beta"}}), true)
			validateContractValue(t, id, input(map[string]any{"project": "alpha"}, map[string]any{"project": []any{"=beta"}}), false)
			validateContractValue(t, id, input(map[string]any{}, map[string]any{"project": []any{"server@=alpha"}}), family == "conditions")
		}
	}
}

func TestConditionMutationInputSchemasSupportPositionalAndBulkForms(t *testing.T) {
	addID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:conditions.add:input"
	add := func(arguments, options map[string]any) map[string]any {
		options["expression"] = "percent <= 1"
		return map[string]any{"arguments": arguments, "options": options, "stdin": nil}
	}
	validateContractValue(t, addID, add(map[string]any{"project": "demo", "name": "Beta"}, map[string]any{}), true)
	validateContractValue(t, addID, add(map[string]any{"name": "Beta"}, map[string]any{}), true)
	validateContractValue(t, addID, add(map[string]any{"name": "Beta"}, map[string]any{"project": []any{"=alpha", "=beta"}}), true)
	validateContractValue(t, addID, add(map[string]any{"name": "Beta"}, map[string]any{"stateless": true, "project": []any{"=alpha", "=beta"}}), true)
	validateContractValue(t, addID, add(map[string]any{"project": "demo", "name": "Beta"}, map[string]any{"project": []any{"=alpha"}}), false)

	deleteID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:conditions.delete:input"
	deleteInput := func(arguments, options map[string]any) map[string]any {
		return map[string]any{"arguments": arguments, "options": options, "stdin": nil}
	}
	validateContractValue(t, deleteID, deleteInput(map[string]any{"project": "demo", "condition": "Beta"}, map[string]any{}), true)
	validateContractValue(t, deleteID, deleteInput(map[string]any{"project": "demo"}, map[string]any{"expr": "usage_count == 0"}), true)
	validateContractValue(t, deleteID, deleteInput(map[string]any{}, map[string]any{"project": []any{"=alpha", "=beta"}, "filter": []any{"=Beta"}}), true)
	validateContractValue(t, deleteID, deleteInput(map[string]any{}, map[string]any{"stateless": true, "project": []any{"=alpha", "=beta"}, "filter": []any{"=Beta"}}), true)
	validateContractValue(t, deleteID, deleteInput(map[string]any{}, map[string]any{"expr": `expression == "true"`}), true)
	validateContractValue(t, deleteID, deleteInput(map[string]any{"condition": "Beta"}, map[string]any{}), false)
	validateContractValue(t, deleteID, deleteInput(map[string]any{"project": "demo", "condition": "Beta"}, map[string]any{"filter": []any{"=Beta"}}), false)
	validateContractValue(t, deleteID, deleteInput(map[string]any{"project": "demo"}, map[string]any{"project": []any{"=alpha"}}), false)
}

func TestStatelessMultiProjectLists(t *testing.T) {
	for _, family := range []string{"apps", "conditions"} {
		for _, mode := range []string{"exact", "discovery", "filtered"} {
			t.Run(family+"/"+mode, func(t *testing.T) {
				var mu sync.Mutex
				requests := map[string]int{}
				cmd, out, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
					path := req.URL.Path
					mu.Lock()
					requests[path]++
					mu.Unlock()
					var body string
					switch {
					case path == "/v1/projects":
						body = `{"projects":[{"projectId":"alpha","projectNumber":"123","lifecycleState":"ACTIVE"},{"projectId":"beta","projectNumber":"456","lifecycleState":"ACTIVE"}]}`
					case strings.HasPrefix(path, "/v3/projects/"):
						id := strings.TrimPrefix(path, "/v3/projects/")
						body = fmt.Sprintf(`{"projectId":%q,"displayName":%q,"state":"ACTIVE"}`, id, strings.ToUpper(id))
					case strings.HasSuffix(path, ":searchApps"):
						body = `{"apps":[{"name":"projects/123/webApps/abc","appId":"1:123:web:abc","displayName":"Web","platform":"WEB","state":"ACTIVE"}]}`
					case strings.HasSuffix(path, "/remoteConfig:listVersions"):
						body = `{"versions":[{"versionNumber":"7"}]}`
					case strings.HasSuffix(path, "/remoteConfig"):
						body = `{"conditions":[{"name":"first","expression":"true"}],"parameters":{},"version":{"versionNumber":"7"}}`
					default:
						t.Fatalf("unexpected request %s", req.URL)
					}
					return statelessHTTPResponse(req, 200, body, `"etag-7"`), nil
				})
				args := []string{"--stateless", family, "list", "--json"}
				switch mode {
				case "exact":
					args = append(args, "-p", "=beta", "-p", "=alpha", "-p", "=alpha")
				case "filtered":
					args = append(args, "-p", "^ALP")
				}
				cmd.SetArgs(args)
				executed, err := cmd.ExecuteC()
				if err != nil {
					t.Fatal(err)
				}
				envelope := contract.BuildEnvelope(executed, "test", out.Bytes(), nil)
				if envelope.Context.Profile != nil || envelope.ExitCode != 0 {
					t.Fatalf("envelope=%#v", envelope)
				}
				raw, err := json.Marshal(envelope)
				if err != nil {
					t.Fatal(err)
				}
				var value map[string]any
				if err := json.Unmarshal(raw, &value); err != nil {
					t.Fatal(err)
				}
				validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:"+family+".list:response", value, true)
				data := value["data"].(map[string]any)
				want := 2
				if mode == "filtered" {
					want = 1
				}
				if data["count"] != float64(want) {
					t.Fatalf("data=%#v", data)
				}
				first := data["items"].([]any)[0].(map[string]any)
				if first["project_id"] != "alpha" {
					t.Fatalf("order=%#v", data)
				}
				wantName := "ALPHA"
				if mode == "exact" {
					wantName = "alpha"
				}
				if first["project"] != wantName {
					t.Fatalf("project name=%#v", first)
				}
				discoveryCount := 1
				if mode == "exact" {
					discoveryCount = 0
				}
				if requests["/v1/projects"] != discoveryCount {
					t.Fatalf("requests=%#v", requests)
				}
				for path, count := range requests {
					if count != 1 {
						t.Fatalf("repeated read %s: %d", path, count)
					}
					if mode == "filtered" && strings.Contains(path, "/beta") && !strings.HasPrefix(path, "/v3/") {
						t.Fatalf("read unselected project %s", path)
					}
				}
				assertProfilePathsAbsent(t, configRoot, cacheRoot)
			})
		}
	}
}
