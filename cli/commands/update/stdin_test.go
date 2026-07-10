package updatecmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
)

func TestRunUpdateStdinPreservesParameterOrder(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	var errOut bytes.Buffer
	raw := `{"version":{"versionNumber":"7"},"parameters":{"b":{"defaultValue":{"value":"old-b"},"valueType":"STRING"},"a":{"defaultValue":{"value":"old-a"},"valueType":"STRING"}}}`
	cmd.SetIn(strings.NewReader(raw))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	spec := updateSpec{
		value: &valueSpec{value: "new-b", valueType: "STRING"},
	}
	err := runUpdateStdin(cmd, []string{"=b"}, "", shared.ParameterSearch{}, spec)
	if err != nil {
		t.Fatalf("runUpdateStdin returned error: %v", err)
	}

	text := out.String()
	bIdx := strings.Index(text, `"b"`)
	aIdx := strings.Index(text, `"a"`)
	if bIdx < 0 || aIdx < 0 {
		t.Fatalf("output missing expected parameter keys:\n%s", text)
	}
	if bIdx > aIdx {
		t.Fatalf("parameter order was not preserved; want b before a:\n%s", text)
	}
	if !strings.Contains(text, `"new-b"`) {
		t.Fatalf("output missing updated value:\n%s", text)
	}
}
