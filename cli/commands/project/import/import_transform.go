package importpkg

import (
	"strings"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
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
	return importer.Transform(project.ProjectID, project.Name, cfg, opts.plannerOptions())
}
