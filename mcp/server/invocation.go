package mcpserver

import (
	"github.com/yumauri/fbrcm/ops"
)

// Invocation retains the published tool input shape. Structured values go
// directly to application operations; they are never encoded as command argv.
type Invocation = ops.Input
