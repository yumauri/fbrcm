package app

import (
	"bytes"
	"github.com/yumauri/fbrcm/ops/contract"
	"strings"
	"testing"
)

func TestExplainIsPublishedForHumanAndMachineUse(t *testing.T) {
	root := NewRootForContract("test")
	cmd, remaining, err := root.Find([]string{"explain"})
	if err != nil || len(remaining) != 0 || cmd.Name() != "explain" {
		t.Fatal("explain missing from root")
	}
	for _, c := range contract.Capabilities(root).Commands {
		if c.ID == "explain" {
			return
		}
	}
	t.Fatal("explain missing from capability discovery")
}

func TestHumanHelpIncludesExplain(t *testing.T) {
	root := NewRootForContract("test")
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil || !strings.Contains(out.String(), "explain") {
		t.Fatalf("help: %v\n%s", err, out.String())
	}
}
