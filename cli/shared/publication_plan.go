package shared

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/rc/publication"
)

// ReadPublicationPlan reads a plan from a path or from stdin when path is "-".
func ReadPublicationPlan(cmd *cobra.Command, path string) (*publication.Plan, error) {
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
