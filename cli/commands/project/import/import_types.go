package importpkg

import (
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/rc/importer"
)

type importOptions struct {
	groups                     []string
	paramFilters               []string
	search                     shared.ParameterSearch
	expr                       string
	removeAllConditions        bool
	keepPortableConditionsOnly bool
	merge                      bool
	override                   bool
	mergeResolve               string
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

func (o importOptions) plannerOptions() importer.Options {
	policy := importer.ConditionPolicyKeep
	if o.removeAllConditions {
		policy = importer.ConditionPolicyRemoveAll
	} else if o.keepPortableConditionsOnly {
		policy = importer.ConditionPolicyKeepPortableOnly
	}
	strategy := importer.StrategyMerge
	if o.override {
		strategy = importer.StrategyReplace
	}
	resolution := importer.ResolutionCurrent
	if o.mergeResolve == string(conflictResolutionImport) {
		resolution = importer.ResolutionImport
	}
	return importer.Options{
		Groups:            append([]string(nil), o.groups...),
		Filters:           append([]string(nil), o.paramFilters...),
		Search:            o.search.Raw,
		Expr:              o.expr,
		Strategy:          strategy,
		ConditionPolicy:   policy,
		DefaultResolution: resolution,
	}
}
