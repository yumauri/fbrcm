// Package fileoutput writes application payloads to private destination files.
package fileoutput

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yumauri/fbrcm/core/config"
)

// Write replaces path with data using private file permissions.
func Write(path string, data []byte) error {
	return write(path, data, false)
}

// Create writes data without replacing an existing destination.
func Create(path string, data []byte) error {
	return write(path, data, true)
}

func write(path string, data []byte, exclusive bool) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create destination dir: %w", err)
		}
	}
	writeFile := config.WritePrivateFile
	if exclusive {
		writeFile = config.WritePrivateFileExclusive
	}
	if err := writeFile(path, data); err != nil {
		return fmt.Errorf("write destination file: %w", err)
	}
	return nil
}
