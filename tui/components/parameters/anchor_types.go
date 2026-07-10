package parameters

import "github.com/yumauri/fbrcm/core"

type RenameAnchor struct {
	Project                   core.Project
	IsGroup                   bool
	GroupKey, ParamKey, Label string
	X, Y, Width, MaxWidth     int
}

type MoveAnchor struct {
	Project                   core.Project
	IsGroup                   bool
	GroupKey, ParamKey, Label string
	X, Y                      int
	Options                   []MoveOption
}

type MoveOption struct{ Key, Label string }

// ConditionalValueAnchor holds selected conditional value deletion target.
type ConditionalValueAnchor struct {
	Project                        core.Project
	GroupKey, ParamKey, ValueLabel string
}

type BoolValueAnchor struct {
	Project                        core.Project
	GroupKey, ParamKey, ValueLabel string
	Value                          bool
	CurrentValue                   string
	X, Y                           int
}

type NumberValueAnchor struct {
	Project                                      core.Project
	GroupKey, ParamKey, ValueLabel, CurrentValue string
	X, Y, Width, MaxWidth                        int
}

type StringValueAnchor struct {
	Project                                      core.Project
	GroupKey, ParamKey, ValueLabel, CurrentValue string
	X, Y, Width, MaxWidth                        int
	FullWidth, Expanded                          bool
}

type JSONValueAnchor struct {
	Project                                      core.Project
	GroupKey, ParamKey, ValueLabel, CurrentValue string
}
