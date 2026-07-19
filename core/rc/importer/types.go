// Package importer implements UI-neutral Remote Config import planning.
package importer

import (
	"encoding/json"
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
)

type Strategy string

const (
	StrategyMerge   Strategy = "merge"
	StrategyReplace Strategy = "replace"
)

type Resolution string

const (
	ResolutionCurrent Resolution = "current"
	ResolutionImport  Resolution = "import"
)

type ConditionPolicy string

const (
	ConditionPolicyKeep             ConditionPolicy = "keep"
	ConditionPolicyKeepPortableOnly ConditionPolicy = "keep_portable_only"
	ConditionPolicyRemoveAll        ConditionPolicy = "remove_all"
)

type Options struct {
	Groups            []string
	Filters           []string
	Search            string
	Expr              string
	Strategy          Strategy
	ConditionPolicy   ConditionPolicy
	DefaultResolution Resolution
	Resolutions       map[string]Resolution
}

type ConflictKind string

const (
	ConflictCondition        ConflictKind = "condition"
	ConflictGroupDescription ConflictKind = "group_description"
	ConflictParameter        ConflictKind = "parameter"
)

type Conflict struct {
	ID      string
	Kind    ConflictKind
	Label   string
	Current any
	Import  any
}

type Summary struct {
	RootParameters        int
	GroupParameters       int
	Groups                int
	Conditions            int
	NonPortableConditions int
	WrappedCache          bool
}

func (s Summary) Parameters() int { return s.RootParameters + s.GroupParameters }

func (s Summary) PortableConditions() int {
	return max(s.Conditions-s.NonPortableConditions, 0)
}

type Plan struct {
	Imported  *firebase.RemoteConfig
	Final     *firebase.RemoteConfig
	Conflicts []Conflict
	Summary   Summary
}

type MissingGroupsError struct {
	Missing   []string
	Available []GroupSummary
}

func (e *MissingGroupsError) Error() string {
	if len(e.Available) == 0 {
		return fmt.Sprintf("requested groups not found in import: %s", joinNames(e.Missing))
	}
	available := make([]string, 0, len(e.Available))
	for _, group := range e.Available {
		available = append(available, group.Name)
	}
	return fmt.Sprintf("requested groups not found in import: %s; available groups: %s", joinNames(e.Missing), joinNames(available))
}

type GroupSummary struct {
	Name       string
	Parameters int
}

type ParsedSource struct {
	Config       *firebase.RemoteConfig
	Raw          json.RawMessage
	WrappedCache bool
}
