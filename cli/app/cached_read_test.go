package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestRemoteConfigCachedFlagContracts(t *testing.T) {
	for _, id := range []string{"get", "conditions.list", "conditions.show", "groups.list", "experiments.list", "experiments.show", "rollouts.list", "rollouts.show", "personalizations.list", "personalizations.show"} {
		t.Run(id, func(t *testing.T) {
			t.Setenv(env.GoogleAccessToken, "test-token")
			root := NewRootForContract("test")
			cmd, _, err := root.Find(strings.Split(id, "."))
			if err != nil || cmd.Flags().Lookup("cached") == nil {
				t.Fatalf("missing --cached: %v", err)
			}
			args := map[string]any{}
			argv := strings.Split(id, ".")
			if id != "get" && id != "groups.list" {
				args["project"] = "demo"
				argv = append(argv, "demo")
			}
			for _, capability := range contract.DetailedCapabilities(root) {
				if capability.ID != id {
					continue
				}
				for _, argument := range capability.Arguments {
					if argument.Required && argument.Name != "project" {
						args[argument.Name] = "example"
						argv = append(argv, "example")
					}
				}
			}
			input := func(options map[string]any) map[string]any {
				return map[string]any{"arguments": args, "options": options, "stdin": nil}
			}
			schema := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + id + ":input"
			validateContractValue(t, schema, input(map[string]any{"cached": true}), true)
			validateContractValue(t, schema, input(map[string]any{"cached": true, "update": true}), false)
			validateContractValue(t, schema, input(map[string]any{"cached": true, "stateless": true}), false)
			for _, flags := range [][]string{{"--cached", "--update", "--json"}, {"--cached", "--stateless", "--json"}} {
				envelope, raw := executeJSONContract(t, append(append([]string{}, argv...), flags...)...)
				if envelope.Outcome != "failure" || envelope.ExitCode != 2 {
					t.Fatalf("invalid combination: %s", raw)
				}
				validateContractDocument(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:"+id+":response", raw)
			}
			if id == "get" {
				value := input(map[string]any{"cached": true})
				value["stdin"] = map[string]any{"version": map[string]any{"versionNumber": "1"}, "parameters": map[string]any{}}
				validateContractValue(t, schema, value, false)
			}
		})
	}
}

func TestCachedRemoteConfigCommandsServeStaleTemplates(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(dir, "config"))
	t.Setenv(env.CacheDir, filepath.Join(dir, "cache"))
	t.Setenv(env.Profile, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride(""); shared.SetMachineMode(false) })
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveAuth(&config.AuthFile{Version: config.AuthConfigVersion, DefaultAuthID: "main", Auth: []config.AuthEntry{config.DefaultOAuthAuthEntry("main", "Main")}}); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"version":{"versionNumber":"7"},"conditions":[{"name":"staff","expression":"true"}],"parameters":{"flag":{"defaultValue":{"personalizationValue":{"personalizationId":"p1"}}}},"parameterGroups":{"empty":{"description":"preserved"}}}`)
	if err := config.SaveParametersCache("demo", &config.ParametersCache{CachedAt: time.Now().Add(-time.Hour), ETag: "etag-7", RemoteConfig: raw}); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"get"}, {"conditions", "list", "demo"}, {"conditions", "show", "demo", "staff"},
		{"groups", "list"}, {"personalizations", "list", "demo"}, {"personalizations", "show", "demo", "p1"},
		{"get", "--filter", "=absent"}, {"groups", "list", "--filter", "=absent"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(strings.NewReader(""))
			cmd.SetArgs(append(args, "--cached", "--json"))
			executed, err := cmd.ExecuteC()
			if err != nil {
				t.Fatal(err)
			}
			envelope := contract.BuildEnvelope(executed, "test", out.Bytes(), nil)
			if envelope.Outcome != "success" || len(envelope.Warnings) != 0 {
				t.Fatalf("envelope=%+v", envelope)
			}
			validateContractDocument(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:"+envelope.Command+":response", marshalEnvelope(t, envelope))
		})
	}
}
