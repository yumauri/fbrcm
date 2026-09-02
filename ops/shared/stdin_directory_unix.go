//go:build unix

package shared

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// OpenStdinDirectoryFile opens a child relative to a directory supplied as
// stdin without requiring a filesystem path for that directory.
func OpenStdinDirectoryFile(dir *os.File, name string) (*os.File, error) {
	fd, err := unix.Openat(int(dir.Fd()), name, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	child := os.NewFile(uintptr(fd), name)
	if child == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("openat returned invalid file")
	}
	return child, nil
}
