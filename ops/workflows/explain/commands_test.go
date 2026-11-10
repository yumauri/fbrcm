package explain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	cliadapter "github.com/yumauri/fbrcm/cli/operation"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/ops/shared"
)

const requestURL = "https://firebaseremoteconfig.googleapis.com/v1/projects/123456789/namespaces/firebase:fetch?key=secret-key"
const appID = "1:123456789:android:abcdef"

type fakePrompter struct {
	answers map[string]evaluate.Fact
	asked   []string
	chosen  int
	onFact  func(evaluate.Need)
}

func (p *fakePrompter) Fact(need evaluate.Need) (evaluate.Fact, error) {
	if p.onFact != nil {
		p.onFact(need)
	}
	p.asked = append(p.asked, need.Key)
	return p.answers[need.Key], nil
}

func (p *fakePrompter) Exchange(exchanges []evaluate.Exchange) (evaluate.Exchange, error) {
	for _, exchange := range exchanges {
		if exchange.Entry == p.chosen {
			return exchange, nil
		}
	}
	return evaluate.Exchange{}, fmt.Errorf("expected explicit exchange selection")
}

func serviceWithTemplate(t *testing.T, template firebase.RemoteConfig) *core.Core {
	t.Helper()
	t.Setenv(env.ConfigDir, t.TempDir())
	t.Setenv(env.CacheDir, t.TempDir())
	t.Setenv(env.NoColor, "1")
	t.Setenv("COLUMNS", "200")
	t.Chdir(t.TempDir())
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo", ProjectNumber: "123456789", Name: "Demo", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(template)
	if err != nil {
		t.Fatal(err)
	}
	if err := config.SaveParametersCache("demo", &config.ParametersCache{CachedAt: time.Now(), RemoteConfig: raw}); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func execute(t *testing.T, svc *core.Core, input string, prompts *fakePrompter, args ...string) (string, error) {
	t.Helper()
	cmd := cliadapter.Command(NewDefinition(svc))
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return run(cmd, svc, args[0], prompts) }
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetIn(shared.DocumentInput(strings.NewReader(input)))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func simpleTemplate() firebase.RemoteConfig {
	return firebase.RemoteConfig{Version: firebase.RemoteConfigVersion{VersionNumber: "142"}, Parameters: map[string]firebase.RemoteConfigParam{
		"flag": {ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}},
	}}
}

func TestExplainStaircasePreservesActualPrioritiesAndSkipsAnalytics(t *testing.T) {
	template := simpleTemplate()
	for i := 1; i <= 31; i++ {
		template.Conditions = append(template.Conditions, firebase.RemoteConfigCondition{Name: fmt.Sprintf("unused_%d", i), Expression: "true"})
	}
	for _, condition := range []struct {
		priority                int
		name, expression, value string
	}{
		{8, "ios_beta", `app.id == '1:123456789:ios:abcdef' && app.audiences.inAtLeastOne(['Beta testers'])`, "true"},
		{16, "android_v4", `app.id == '` + appID + `' && app.version.exactlyMatches(['4.0.0'])`, "true"},
		{23, "android_english", `app.id == '` + appID + `' && device.language in ['en-US']`, "true"},
		{31, "android_users", `app.id == '` + appID + `'`, "false"},
	} {
		template.Conditions[condition.priority-1] = firebase.RemoteConfigCondition{Name: condition.name, Expression: condition.expression, TagColor: "GREEN"}
		param := template.Parameters["flag"]
		if param.ConditionalValues == nil {
			param.ConditionalValues = make(map[string]firebase.RemoteConfigValue)
		}
		param.ConditionalValues[condition.name] = firebase.RemoteConfigValue{Value: condition.value}
		template.Parameters["flag"] = param
	}
	// Grouped parameters follow exactly the same lookup and resolution path.
	template.ParameterGroups = map[string]firebase.RemoteConfigGroup{"players": {Parameters: template.Parameters}}
	template.Parameters = nil
	svc := serviceWithTemplate(t, template)
	input := `curl '` + requestURL + `' --data-raw '{"appId":"` + appID + `","appVersion":"3.8.0","languageCode":"en-US","appInstanceIdToken":"secret-token"}'`
	prompts := &fakePrompter{}
	out, err := execute(t, svc, input, prompts, "flag", "--cached")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"× [8] ios_beta", "× [16] android_v4", "✓ [23] android_english", "✓ [31] android_users", "◆ flag = true", "Winner: [23] android_english", "├─ [31] android_users also matches, but lower priority", "Default false is not used"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Left side") || strings.Contains(out, "skip the rest") {
		t.Fatalf("printed unnecessary control-flow steps:\n%s", out)
	}
	if len(prompts.asked) != 0 || strings.Contains(out, "secret") || strings.Contains(out, "unused_") {
		t.Fatalf("unnecessary question or leaked unrelated data: %#v\n%s", prompts.asked, out)
	}
}

func TestExplainMissingParameterAndDefaultDoNotAsk(t *testing.T) {
	svc := serviceWithTemplate(t, simpleTemplate())
	for _, key := range []string{"flag", "missing"} {
		prompts := &fakePrompter{}
		// No capture is required when project and parameter suffice.
		cmd := cliadapter.Command(NewDefinition(svc))
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs([]string{key, "--project", "demo", "--cached"})
		cmd.SetIn(strings.NewReader(""))
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return run(cmd, svc, args[0], prompts) }
		err := cmd.Execute()
		if key == "missing" {
			if err == nil || !strings.Contains(err.Error(), `parameter "missing" was not found`) {
				t.Fatalf("missing parameter = %v", err)
			}
		} else if err != nil || !strings.Contains(out.String(), "◆ flag = false") || strings.Contains(out.String(), "eval:") {
			t.Fatalf("default-only result: %v\n%s", err, out.String())
		}
		if len(prompts.asked) != 0 {
			t.Fatalf("asked for unused context: %#v", prompts.asked)
		}
	}
}

func TestUnknownHigherPriorityPreventsFalseWinnerAndInputIsMemoized(t *testing.T) {
	template := simpleTemplate()
	template.Conditions = []firebase.RemoteConfigCondition{
		{Name: "premium", Expression: `app.userProperty['tier'].exactlyMatches(['paid'])`},
		{Name: "staff", Expression: `app.userProperty['tier'].exactlyMatches(['staff'])`},
		{Name: "everyone", Expression: "true"},
	}
	param := template.Parameters["flag"]
	param.ConditionalValues = map[string]firebase.RemoteConfigValue{"premium": {Value: "true"}, "staff": {Value: "true"}, "everyone": {Value: "false"}}
	template.Parameters["flag"] = param
	svc := serviceWithTemplate(t, template)
	input := `curl '` + requestURL + `' --data '{}'`
	prompts := &fakePrompter{}
	out, err := execute(t, svc, input, prompts, "flag", "--cached")
	if err != nil || !strings.Contains(out, "? flag = unknown") || strings.Contains(out, "Winner:") || len(prompts.asked) != 1 {
		t.Fatalf("unknown result: %v, prompts %#v\n%s", err, prompts.asked, out)
	}
	for _, want := range []string{
		"├─ If app.userProperty['tier'].exactlyMatches(['paid']), then [1] premium would win → true",
		"├─ Otherwise, if app.userProperty['tier'].exactlyMatches(['staff']), then [2] staff would win → true",
		"╰─ Otherwise: [3] everyone would win → false",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing outcome %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Otherwise: default") {
		t.Fatalf("default is unreachable when everyone matches:\n%s", out)
	}
	out, err = execute(t, svc, input, &fakePrompter{}, "flag", "--cached", "--no-input", "--set", `app.userProperty["tier"]=paid`)
	if err != nil || !strings.Contains(out, "Winner: [1] premium") || !strings.Contains(out, "command line") {
		t.Fatalf("supplied input result: %v\n%s", err, out)
	}
}

func TestProjectTemplateAndRequestArePrintedBeforeMissingInputPrompts(t *testing.T) {
	template := conditionTemplate(`app.userProperty['tier'].exactlyMatches(['paid']) && app.userProperty['role'].exactlyMatches(['staff'])`)
	svc := serviceWithTemplate(t, template)
	var out bytes.Buffer
	header := "demo / flag\nTemplate v142 · cache\nRequest: stdin · entry 1\n"
	prompts := &fakePrompter{answers: map[string]evaluate.Fact{
		"app.userProperty['tier']": {Known: true, Value: "paid", Source: "your answer"},
		"app.userProperty['role']": {Known: true, Value: "staff", Source: "your answer"},
	}, onFact: func(need evaluate.Need) {
		if !strings.HasPrefix(out.String(), header) {
			t.Fatalf("missing context before asking for %s:\n%s", need.Key, out.String())
		}
		if strings.Contains(out.String(), "eval:") {
			t.Fatalf("evaluation printed before all prompts finished:\n%s", out.String())
		}
		_, _ = fmt.Fprintln(&out, "Prompt: "+need.Key)
	}}
	cmd := cliadapter.Command(NewDefinition(svc))
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return run(cmd, svc, args[0], prompts) }
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetIn(shared.DocumentInput(strings.NewReader(`curl '` + requestURL + `' --data '{}'`)))
	cmd.SetArgs([]string{"flag", "--cached"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(prompts.asked) != 2 || strings.Count(out.String(), header) != 1 || !strings.Contains(out.String(), "◆ flag = true") {
		t.Fatalf("context was duplicated or evaluation did not follow prompts: %v\n%s", prompts.asked, out.String())
	}
}

func TestCaptureAutodetectionForFileAndStdin(t *testing.T) {
	svc := serviceWithTemplate(t, simpleTemplate())
	curl := `curl '` + requestURL + `' --data '{}'`
	har := `{"log":{"entries":[{"request":{"method":"POST","url":"` + requestURL + `","postData":{"text":"{}"}},"response":{"status":200,"content":{"text":"{\"entries\":{\"flag\":\"true\"},\"templateVersion\":\"141\"}"}}}]}}`
	for _, input := range []string{curl, har} {
		for _, file := range []bool{false, true} {
			args := []string{"flag", "--cached"}
			stdin := input
			if file {
				path := filepath.Join(t.TempDir(), "capture.unknown-extension")
				if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
					t.Fatal(err)
				}
				args, stdin = append(args, "--from", path), ""
			} else {
				args = append(args, "--from", "-")
			}
			out, err := execute(t, svc, stdin, &fakePrompter{}, args...)
			if err != nil || !strings.Contains(out, "◆ flag = false") {
				t.Fatalf("input detection (file %t): %v\n%s", file, err, out)
			}
			if input == har && (!strings.Contains(out, "Captured response: true") || !strings.Contains(out, "Capture template v141 differs")) {
				t.Fatalf("response evidence missing:\n%s", out)
			}
		}
	}
}

func TestSelectsOriginalHARFetchEntryAndRejectsWrongProject(t *testing.T) {
	svc := serviceWithTemplate(t, simpleTemplate())
	har := `{"log":{"entries":[{"request":{"method":"GET","url":"https://example.com"}},{"request":{"method":"POST","url":"` + requestURL + `","postData":{"text":"{}"}}},{"request":{"method":"POST","url":"` + requestURL + `","postData":{"text":"{}"}}}]}}`
	out, err := execute(t, svc, har, &fakePrompter{chosen: 3}, "flag", "--cached")
	if err != nil || !strings.Contains(out, "entry 3") {
		t.Fatalf("entry selection = %v\n%s", err, out)
	}
	if _, err := execute(t, svc, har, &fakePrompter{}, "flag", "--cached", "--no-input"); err == nil || !strings.Contains(err.Error(), "--entry") {
		t.Fatalf("ambiguous archive silently selected a fetch: %v", err)
	}
	if _, err := execute(t, svc, har, &fakePrompter{}, "flag", "--cached", "--entry", "1"); err == nil {
		t.Fatal("accepted non-fetch HAR entry")
	}
	wrong := strings.ReplaceAll(har, "/projects/123456789/", "/projects/987654321/")
	if _, err := execute(t, svc, wrong, &fakePrompter{}, "flag", "--cached", "--project", "demo", "--entry", "2"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("accepted mismatched project: %v", err)
	}
}

func TestJSONCaptureFileFailurePrecedesTemplateReads(t *testing.T) {
	root := &cobra.Command{Use: "fbrcm", SilenceUsage: true}
	root.PersistentFlags().Bool("json", false, "")
	root.AddCommand(cliadapter.Command(NewDefinition(nil)))
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"explain", "flag", "--json", "--from", "nonexistent"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "open capture") || out.Len() != 0 {
		t.Fatalf("JSON boundary = %v, %q", err, out.String())
	}
}

func platformTemplate() firebase.RemoteConfig {
	template := simpleTemplate()
	param := template.Parameters["flag"]
	param.ConditionalValues = make(map[string]firebase.RemoteConfigValue)
	for _, platform := range []string{"ios", "android", "web"} {
		template.Conditions = append(template.Conditions, firebase.RemoteConfigCondition{Name: platform, Expression: "device.os == '" + platform + "'"})
		param.ConditionalValues[platform] = firebase.RemoteConfigValue{Value: "true"}
	}
	template.Parameters["flag"] = param
	return template
}

func TestPlatformIsInferredFromCapturedAppIDWithoutPrompting(t *testing.T) {
	svc := serviceWithTemplate(t, platformTemplate())
	for priority, platform := range []string{"ios", "android", "web"} {
		for _, field := range []string{"app_id", "appId"} {
			body := fmt.Sprintf(`{"%s":"1:123456789:%s:abcdef"}`, field, platform)
			for _, format := range []string{"curl", "har"} {
				t.Run(platform+"/"+field+"/"+format, func(t *testing.T) {
					input := "curl '" + requestURL + "' --data-raw '" + body + "'"
					if format == "har" {
						input = fmt.Sprintf(`{"log":{"entries":[{"request":{"method":"POST","url":%q,"postData":{"text":%q}}}]}}`, requestURL, body)
					}
					prompts := &fakePrompter{}
					out, err := execute(t, svc, input, prompts, "flag", "--cached")
					winner := fmt.Sprintf("Winner: [%d] %s", priority+1, platform)
					if err != nil || !strings.Contains(out, winner) || !strings.Contains(out, "device.os = '"+platform+"' (request app id)") || len(prompts.asked) != 0 {
						t.Fatalf("platform inference: %v, prompts %v\n%s", err, prompts.asked, out)
					}
				})
			}
		}
	}
}

func TestExplicitInputsTakePrecedenceOverCapturedPlatform(t *testing.T) {
	svc := serviceWithTemplate(t, platformTemplate())
	input := `curl '` + requestURL + `' --data-raw '{"app_id":"1:123456789:web:abcdef"}'`
	for _, test := range []struct {
		name   string
		inputs []string
		want   string
	}{
		{"platform override", []string{"--set", "device.os=android"}, "Winner: [2] android"},
		{"effective App ID override", []string{"--set", "app.id=1:123456789:ios:abcdef"}, "Winner: [1] ios"},
		{"platform overrides both App IDs", []string{"--set", "app.id=1:123456789:ios:abcdef", "--set", "device.os=android"}, "Winner: [2] android"},
		{"explicitly unset platform", []string{"--set", "device.os=/unset"}, "◆ flag = false"},
		{"explicitly unknown platform", []string{"--set", "device.os=/unknown"}, "? flag = unknown"},
	} {
		t.Run(test.name, func(t *testing.T) {
			prompts := &fakePrompter{}
			args := append([]string{"flag", "--cached"}, test.inputs...)
			out, err := execute(t, svc, input, prompts, args...)
			if err != nil || !strings.Contains(out, test.want) || len(prompts.asked) != 0 {
				t.Fatalf("explicit input precedence: %v, prompts %v\n%s", err, prompts.asked, out)
			}
			if test.name == "effective App ID override" && !strings.Contains(out, "command line app id") {
				t.Fatalf("derived platform lost its input provenance:\n%s", out)
			}
		})
	}
}

func TestMissingOrInvalidAppIDStillAsksForPlatform(t *testing.T) {
	svc := serviceWithTemplate(t, platformTemplate())
	for _, body := range []string{`{}`, `{"app_id":null}`, `{"app_id":"invalid"}`, `{"app_id":"1:123456789:desktop:abcdef"}`} {
		input := "curl '" + requestURL + "' --data-raw '" + body + "'"
		prompts := &fakePrompter{answers: map[string]evaluate.Fact{"device.os": {Known: true, Value: "web", Source: "your answer"}}}
		out, err := execute(t, svc, input, prompts, "flag", "--cached")
		if err != nil || !strings.Contains(out, "Winner: [3] web") || len(prompts.asked) != 1 || prompts.asked[0] != "device.os" {
			t.Fatalf("missing platform fallback: %v, prompts %v\n%s", err, prompts.asked, out)
		}
	}
}
