// Package rootgroup defines the canonical representations of the Remote Config
// "root group" (the default, ungrouped parameter bucket) and adapters between
// them. Firebase stores default parameters outside of parameterGroups, so the
// root group has no real name on the wire. fbrcm uses three distinct
// representations depending on the layer; they are collected here so the values
// and conversions live in one place.
//
// See docs/root-group-key.md for the rationale.
package rootgroup

const (
	// WireKey is the root group key on the Firebase wire and in draft slots:
	// the default bucket is the implicit empty-named group.
	WireKey = ""

	// TreeKey is the synthetic node key used by the parameters tree and TUI so
	// the default bucket has a stable identity distinct from real groups.
	TreeKey = "__default__"

	// Label is the human-facing label for the root group, also used as the
	// group value in filter expressions.
	Label = "(root)"
)
