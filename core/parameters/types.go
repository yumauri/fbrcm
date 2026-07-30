package parameters

import (
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
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
	Label           string
	Value           string
	Display         rcdisplay.ValueSummary
	RawValue        string
	ValueType       string
	Color           string
	Empty           bool
	EmptyType       string
	Plain           bool
	UseInAppDefault bool
}

// ReadOnly reports whether Firebase or an unsupported future union option owns
// this value rather than fbrcm's plain-value editors.
func (v Value) ReadOnly() bool {
	return v.Display.Kind != rcdisplay.ValueSummaryPlain
}

// HasReadOnlyValues reports whether any value in the parameter is opaque to
// fbrcm mutations.
func (e Entry) HasReadOnlyValues() bool {
	for _, value := range e.Values {
		if value.ReadOnly() {
			return true
		}
	}
	return false
}

// HasReadOnlyValues reports whether any parameter in the group contains an
// opaque value.
func (g Group) HasReadOnlyValues() bool {
	for _, parameter := range g.Parameters {
		if parameter.HasReadOnlyValues() {
			return true
		}
	}
	return false
}
