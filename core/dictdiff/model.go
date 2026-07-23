// Package dictdiff compares generic named dictionaries for reusable diff
// renderers. It owns normalization and contextual line alignment, but no
// terminal styling or layout.
package dictdiff

import "encoding/json"

type ValueType string

const (
	ValueString  ValueType = "string"
	ValueBoolean ValueType = "boolean"
	ValueNumber  ValueType = "number"
	ValueJSON    ValueType = "json"
	ValueNull    ValueType = "null"
)

// ComparisonHint controls how a value is normalized and compared.
type ComparisonHint string

const (
	// CompareString performs a regular textual diff.
	CompareString ComparisonHint = "string"
	// CompareJSON prettifies valid JSON before performing a textual diff. Invalid
	// JSON falls back to its raw string representation.
	CompareJSON ComparisonHint = "json"
	// CompareEnum treats each value as an indivisible choice.
	CompareEnum ComparisonHint = "enum"
)

// Value is one explicitly typed dictionary value. Raw contains the string,
// number, or JSON representation; booleans use Boolean; null has no payload.
type Value struct {
	Type      ValueType
	CompareAs ComparisonHint
	Raw       string
	Boolean   bool
}

func String(value string) Value {
	return Value{Type: ValueString, CompareAs: CompareString, Raw: value}
}
func Enum(value string) Value {
	return Value{Type: ValueString, CompareAs: CompareEnum, Raw: value}
}
func Boolean(value bool) Value {
	return Value{Type: ValueBoolean, CompareAs: CompareEnum, Boolean: value}
}
func Number(value json.Number) Value {
	return Value{Type: ValueNumber, CompareAs: CompareEnum, Raw: value.String()}
}
func JSON(value string) Value {
	return Value{Type: ValueJSON, CompareAs: CompareJSON, Raw: value}
}
func Null() Value {
	return Value{Type: ValueNull, CompareAs: CompareEnum}
}

type Dictionary map[string]Value

type NamedDictionary struct {
	Name       string
	Properties Dictionary
}

type Input struct {
	EntityName string
	Left       NamedDictionary
	Right      NamedDictionary
	// ContextLines controls logical context prepared around each change.
	// Renderers may reduce it further after terminal wrapping.
	ContextLines int
}

type ChangeKind string

const (
	ChangeAdded   ChangeKind = "added"
	ChangeRemoved ChangeKind = "removed"
	ChangeChanged ChangeKind = "changed"
)

type LineKind string

const (
	LineEqual   LineKind = "equal"
	LineAdded   LineKind = "added"
	LineRemoved LineKind = "removed"
	LineChanged LineKind = "changed"
)

type Line struct {
	Text     string
	Segments []Segment
}

type Segment struct {
	Text string
	Kind LineKind
}

type Row struct {
	Left  *Line
	Right *Line
	Kind  LineKind
}

type Chunk struct {
	Rows []Row
}

type PreparedValue struct {
	Type      ValueType
	CompareAs ComparisonHint
	Text      string
}

type Property struct {
	Name   string
	Kind   ChangeKind
	Left   *PreparedValue
	Right  *PreparedValue
	Chunks []Chunk
}

type Result struct {
	EntityName string
	LeftName   string
	RightName  string
	Properties []Property
}
