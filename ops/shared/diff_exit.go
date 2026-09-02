package shared

import (
	"github.com/yumauri/fbrcm/ops/invocation"
)

// DiffFoundError returns status 1 when a successful comparison found
// differences. Both human and JSON output use this semantic result status.
func DiffFoundError(cmd invocation.Call) error {
	return WithExitCode(nil, 1)
}
