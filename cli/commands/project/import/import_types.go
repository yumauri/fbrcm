package importpkg

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/cli/shared"
)

type importOptions struct {
	groups                          []string
	paramFilters                    []string
	search                          shared.ParameterSearch
	expr                            string
	removeAllConditions             bool
	removeProjectSpecificConditions bool
	merge                           bool
	override                        bool
	mergeResolve                    string
}

type importStrategy string

const (
	importStrategyMerge    importStrategy = "merge"
	importStrategyOverride importStrategy = "override"
)

type conflictResolution string

const (
	conflictResolutionCurrent conflictResolution = "current"
	conflictResolutionImport  conflictResolution = "import"
)

type mergeChoice struct {
	label string
	value string
}

func (c mergeChoice) String() string {
	return c.label
}

type missingImportGroupsError struct {
	missing   []string
	available []groupSummary
}

func (e *missingImportGroupsError) Error() string {
	if len(e.available) > 0 {
		available := make([]string, 0, len(e.available))
		for _, group := range e.available {
			available = append(available, group.Name)
		}
		return fmt.Sprintf("requested groups not found in import: %s; available groups: %s", strings.Join(e.missing, ", "), strings.Join(available, ", "))
	}
	return fmt.Sprintf("requested groups not found in import: %s", strings.Join(e.missing, ", "))
}
