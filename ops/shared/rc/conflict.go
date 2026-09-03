package rc

import (
	"errors"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/machine"
)

// IsRemoteConfigConflict reports typed ETag/precondition conflicts.
func IsRemoteConfigConflict(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := errors.AsType[*machine.ConflictError](err); ok {
		return true
	}
	var apiErr *firebase.APIError
	return errors.As(err, &apiErr) && (apiErr.StatusCode == 409 || apiErr.StatusCode == 412)
}
