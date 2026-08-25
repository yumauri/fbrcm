//go:build !unix

package shared

import (
	"fmt"
	"os"
)

// OpenStdinDirectoryFile opens a child relative to a directory supplied as
// stdin without requiring a filesystem path for that directory.
func OpenStdinDirectoryFile(_ *os.File, _ string) (*os.File, error) {
	return nil, fmt.Errorf("directory stdin is unsupported on this platform")
}
