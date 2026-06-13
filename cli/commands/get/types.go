package get

import (
	"image/color"
	"time"
)

type parameterConditionJSON struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
}

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

type parameterRow struct {
	Project      string
	ProjectID    string
	Group        string
	Key          string
	Description  string
	DefaultValue *string
	Conditional  bool
	Conditions   []parameterConditionJSON
	Type         string
	Version      string
	CachedAt     time.Time
	Status       string
	ValueLines   []valueLine
}

type tableLayout struct {
	includeProject bool
	includeGroup   bool
	includeKey     bool
	includeType    bool
	showNames      bool
	valueWidth     int
}

type valueLine struct {
	Label     string
	Value     string
	Color     color.Color
	IsDefault bool
	Missing   bool
	ValueType string
}
