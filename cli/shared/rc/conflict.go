package rc

import (
	"errors"

	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/core/firebase"
)

// IsRemoteConfigConflict reports typed ETag/precondition conflicts.
func IsRemoteConfigConflict(err error) bool {
	if err == nil {
		return false
	}

	var conflict *machine.ConflictError
	if errors.As(err, &conflict) {
		return true
	}
	var apiErr *firebase.APIError
	return errors.As(err, &apiErr) && (apiErr.StatusCode == 409 || apiErr.StatusCode == 412)
}
