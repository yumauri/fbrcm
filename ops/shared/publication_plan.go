package shared

import (
	"fmt"
	"io"
	"os"

	"github.com/yumauri/fbrcm/core/rc/publication"
	"github.com/yumauri/fbrcm/ops/invocation"
)

// ReadPublicationPlan reads a plan from a path or from stdin when path is "-".
func ReadPublicationPlan(cmd invocation.Call, path string) (*publication.Plan, error) {
	var raw []byte
	var err error
	if path == "-" {
		raw, err = io.ReadAll(cmd.InOrStdin())
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("read publication plan: %w", err)
	}
	plan, err := publication.Parse(raw)
	if err != nil {
		return nil, err
	}
	return plan, nil
}
