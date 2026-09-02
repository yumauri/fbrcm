package shared

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core"
)

// RejectStatelessDraft rejects local draft persistence when execution policy
// disables application-managed local state.
func RejectStatelessDraft(ctx context.Context, draft bool) error {
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState || !draft {
		return nil
	}
	return InvalidArgument(fmt.Errorf("--draft cannot be used with --stateless; stateless execution does not write local drafts"))
}
