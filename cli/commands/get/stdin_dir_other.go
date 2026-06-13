//go:build !unix

package get

import (
	"fmt"
	"os"
)

// readStdinDirectoryFile reads a child file relative to a directory passed as stdin.
func readStdinDirectoryFile(_ *os.File, _ string) ([]byte, error) {
	return nil, fmt.Errorf("directory stdin is unsupported on this platform")
}
