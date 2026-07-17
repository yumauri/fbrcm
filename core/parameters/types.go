package parameters

import "time"

type Tree struct {
	Version    string
	CachedAt   time.Time
	ETag       string
	Conditions []Condition
	Groups     []Group
}

type Condition struct {
	Name        string
	Expression  string
	Description string
	Color       string
}

type Group struct {
	Key         string
	Label       string
	Description string
	Parameters  []Entry
}

type Entry struct {
	Key         string
	Description string
	Summary     string
	Values      []Value
}

type Value struct {
	Label     string
	Value     string
	RawValue  string
	ValueType string
	Color     string
	Empty     bool
	EmptyType string
	Plain     bool
}
