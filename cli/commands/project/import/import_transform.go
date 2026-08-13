package importpkg

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func normalizeGroups(groups []string) []string {
	seen := make(map[string]struct{}, len(groups))
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, ok := seen[group]; !ok {
			seen[group] = struct{}{}
			out = append(out, group)
		}
	}
	return out
}

func transformImportConfig(project core.Project, cfg *firebase.RemoteConfig, opts importOptions) error {
	projectID := project.ProjectID
	if target, err := rctarget.Parse(projectID); err == nil {
		projectID = target.ProjectID
	}
	err := importer.Transform(projectID, project.Name, cfg, opts.plannerOptions())
	if err != nil && strings.TrimSpace(opts.expr) != "" {
		return &shared.ExpressionError{Expression: opts.expr, Context: "import_parameter", Target: projectID, Err: fmt.Errorf("transform imported config: %w", err)}
	}
	return err
}
