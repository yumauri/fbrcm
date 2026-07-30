package managedfeatures

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestExperimentShowJSONPreservesWireDataAndBindingPresence(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := coreconfig.SwitchProfile(coreconfig.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	project := core.Project{
		Name: "Demo", ProjectID: "demo", AuthID: "main", DiscoveredBy: []string{"main"},
	}
	if err := coreconfig.SaveProjects([]coreconfig.Project{project}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddGCloudAuth("main", "main"); err != nil {
		t.Fatal(err)
	}

	var requests []string
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: managedFeatureJSONRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.RequestURI())
		var body string
		switch req.URL.Path {
		case "/v1/projects/demo/namespaces/firebase/experiments/exp-1":
			body = `{
				"name":"projects/demo/namespaces/firebase/experiments/exp-1",
				"definition":{"displayName":"Signup","futureNestedField":{"kept":true}},
				"state":"RUNNING",
				"futureTopLevelField":"kept"
			}`
		case "/v1/projects/demo/remoteConfig:listVersions":
			body = `{"versions":[{"versionNumber":"17"}]}`
		case "/v1/projects/demo/remoteConfig":
			body = `{
				"version":{"versionNumber":"17"},
				"parameters":{
					"signup_message":{
						"defaultValue":{"experimentValue":{
							"experimentId":"exp-1",
							"variantValue":[{"variantId":"control","value":""}],
							"exposurePercent":0
						}},
						"valueType":"STRING"
					}
				}
			}`
		default:
			return nil, fmt.Errorf("unexpected request: %s %s", req.Method, req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{"etag-17"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})})
	svc.InjectFirebaseService("main", client)

	cmd := newExperimentsShowCommand(svc)
	cmd.SetArgs([]string{"demo", "exp-1", "--json"})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("decode JSON output: %v\n%s", err, stdout.String())
	}
	experiment := output["experiment"].(map[string]any)
	definition := experiment["definition"].(map[string]any)
	if experiment["futureTopLevelField"] != "kept" || definition["futureNestedField"] == nil {
		t.Fatalf("JSON output dropped beta API fields: %s", stdout.String())
	}
	references := experiment["references"].([]any)
	reference := references[0].(map[string]any)
	if reference["percentage"] != float64(0) {
		t.Fatalf("zero exposure presence was lost: %s", stdout.String())
	}
	variant := reference["variants"].([]any)[0].(map[string]any)
	if value, ok := variant["value"]; !ok || value != "" {
		t.Fatalf("empty variant value presence was lost: %s", stdout.String())
	}
	if len(requests) != 3 ||
		requests[0] != "/v1/projects/demo/namespaces/firebase/experiments/exp-1" ||
		!strings.HasPrefix(requests[1], "/v1/projects/demo/remoteConfig:listVersions") ||
		!strings.HasPrefix(requests[2], "/v1/projects/demo/remoteConfig") {
		t.Fatalf("managed-feature requests = %#v", requests)
	}
}

type managedFeatureJSONRoundTripFunc func(*http.Request) (*http.Response, error)

func (f managedFeatureJSONRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
