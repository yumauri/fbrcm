package parameters

import (
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

type Tree struct {
	Version    string
	CachedAt   time.Time
	ETag       string
	Conditions []Condition
	Groups     []Group

	remoteConfig *firebase.RemoteConfig
}

// RemoteConfig returns the source config used to build the read-only tree.
func (t *Tree) RemoteConfig() *firebase.RemoteConfig {
	if t == nil {
		return nil
	}
	return t.remoteConfig
}

type Condition struct {
	Name       string
	Expression string
	Color      string
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
