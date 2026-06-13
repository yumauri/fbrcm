//go:build unix

package get

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// readStdinDirectoryFile reads a child file relative to a directory passed as stdin.
func readStdinDirectoryFile(dir *os.File, name string) ([]byte, error) {
	fd, err := unix.Openat(int(dir.Fd()), name, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	child := os.NewFile(uintptr(fd), name)
	if child == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("openat returned invalid file")
	}
	defer func() { _ = child.Close() }()

	return io.ReadAll(child)
}
