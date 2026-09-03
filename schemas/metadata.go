package schemas

import (
	_ "embed"
)

// CapabilitiesJSON is generated together with the CLI schemas from the same
// shared operation definitions. Protocol startup does not build a CLI tree.
//
//go:embed capabilities.json
var CapabilitiesJSON []byte
