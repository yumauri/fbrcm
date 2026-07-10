package get

import (
	"time"

	"github.com/yumauri/fbrcm/cli/commands/get/table"
)

type parameterConditionJSON = table.ParameterConditionJSON

type parameterRowJSON struct {
	Project      string                   `json:"project"`
	ProjectID    string                   `json:"project_id"`
	Group        string                   `json:"group"`
	Key          string                   `json:"key"`
	Description  string                   `json:"description"`
	DefaultValue *string                  `json:"default_value"`
	Conditional  bool                     `json:"conditional"`
	Conditions   []parameterConditionJSON `json:"conditions"`
	Type         string                   `json:"type"`
	Version      *string                  `json:"version"`
	CachedAt     *time.Time               `json:"cached_at"`
	Status       *string                  `json:"status"`
}

type parameterRow = table.Row
type valueLine = table.ValueLine
