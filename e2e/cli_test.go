package e2e

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/e2e/internal/harness"
)

var (
	e2eMode = flag.String("mode", "replay", "E2E fixture mode: replay, record-missing, refresh-http, update-output, or refresh-all")
	binary  = flag.String("binary", "", "path to an fbrcm binary; default builds the current checkout")
	goRun   = flag.String("go-run", "", "run fbrcm through 'go run <path>' instead of building it")
	caCert  = flag.String("ca-cert", "", "existing Hoverfly CA certificate; must be paired with -ca-key")
	caKey   = flag.String("ca-key", "", "existing Hoverfly CA private key; must be paired with -ca-cert")
)

func TestCLI(t *testing.T) {
	e2eRoot, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	mode, err := harness.ParseMode(*e2eMode)
	if err != nil {
		t.Fatal(err)
	}
	suite, err := harness.LoadSuite(filepath.Join(e2eRoot, "testdata", "suite.json"))
	if err != nil {
		t.Fatal(err)
	}
	scenarios, err := harness.LoadScenarios(filepath.Join(e2eRoot, "testdata", "scenarios"), suite)
	if err != nil {
		t.Fatal(err)
	}
	runPattern := ""
	if runFlag := flag.Lookup("test.run"); runFlag != nil {
		runPattern = runFlag.Value.String()
	}
	if err := harness.ValidateRecordingRunFilter(suite, mode, runPattern); err != nil {
		t.Fatal(err)
	}
	scenarios, err = harness.OrderScenariosForMode(scenarios, suite, mode)
	if err != nil {
		t.Fatal(err)
	}
	toolDirectory := t.TempDir()
	setupContext, cancelSetup := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelSetup()
	cli, err := harness.ResolveCLI(setupContext, e2eRoot, *binary, *goRun, toolDirectory)
	if err != nil {
		t.Fatal(err)
	}
	hoverfly, err := harness.ResolveHoverfly(setupContext, e2eRoot, toolDirectory)
	if err != nil {
		t.Fatal(err)
	}
	guard, err := harness.BuildReadGuard(setupContext, e2eRoot, toolDirectory)
	if err != nil {
		t.Fatal(err)
	}
	certificate, key := *caCert, *caKey
	if (certificate == "") != (key == "") {
		t.Fatal("-ca-cert and -ca-key must be provided together")
	}
	if certificate == "" {
		certificate, key, err = harness.GenerateCA(setupContext, hoverfly, filepath.Join(toolDirectory, "ca"))
		if err != nil {
			t.Fatal(err)
		}
	} else {
		certificate, err = filepath.Abs(certificate)
		if err != nil {
			t.Fatal(err)
		}
		key, err = filepath.Abs(key)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			report, err := harness.RunScenario(ctx, scenario, suite, harness.RunOptions{
				Mode:              mode,
				CLI:               cli,
				HoverflyBinary:    hoverfly,
				GuardPath:         guard,
				CertificatePath:   certificate,
				KeyPath:           key,
				FixturesRoot:      filepath.Join(e2eRoot, "testdata", "fixtures"),
				StateFixturesRoot: filepath.Join(e2eRoot, "testdata", "state"),
				SchemasRoot:       filepath.Join(e2eRoot, "..", "schemas", "cli", "1.0.0"),
				AccessToken:       os.Getenv("FBRCM_E2E_ACCESS_TOKEN"),
				ScenarioRoot:      t.TempDir(),
				ToolsRoot:         toolDirectory,
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, change := range report.Changes {
				action := "updated"
				if change.Created {
					action = "created"
				}
				t.Logf("%s %s", action, relativePath(e2eRoot, change.Path))
			}
		})
	}
}

func relativePath(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return relative
}
