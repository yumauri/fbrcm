package table

import (
	"image/color"
	"time"
)

// Row is one rendered parameter line in the get table output.
type Row struct {
	Project      string
	ProjectID    string
	Group        string
	Key          string
	Description  string
	DefaultValue *string
	Conditional  bool
	Conditions   []ParameterConditionJSON
	Type         string
	Version      string
	CachedAt     time.Time
	Status       string
	ValueLines   []ValueLine
}

// ValueLine is one value cell line inside a table row.
type ValueLine struct {
	Label     string
	Value     string
	Color     color.Color
	IsDefault bool
	Missing   bool
	ValueType string
}

// ParameterConditionJSON mirrors the JSON shape for parameter conditions.
type ParameterConditionJSON struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
}

type tableLayout struct {
	includeProject bool
	includeGroup   bool
	includeKey     bool
	includeType    bool
	showNames      bool
	valueWidth     int
}
