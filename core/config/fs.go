package config

import (
	"io"
	"os"
	"path/filepath"

	corelog "github.com/yumauri/fbrcm/core/log"
)

const (
	PrivateDirMode  = 0o700
	PrivateFileMode = 0o600
)

// Create/check directory and ensure it is private to the user
func EnsurePrivateDir(path string) error {
	logger := corelog.For("config")
	logger.Debug("ensure private dir", "path", path, "mode", "0700")
	if err := os.MkdirAll(path, PrivateDirMode); err != nil {
		logger.Error("create private dir failed", "path", path, "err", err)
		return err
	}
	if err := os.Chmod(path, PrivateDirMode); err != nil {
		logger.Error("chmod private dir failed", "path", path, "err", err)
		return err
	}
	logger.Debug("private dir ready", "path", path)
	return nil
}

// WritePrivateFile writes data with private file mode and ensures permissions.
func WritePrivateFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, PrivateFileMode); err != nil {
		return err
	}
	return EnsurePrivateFile(path)
}

// WritePrivateFileExclusive creates a private file without replacing an
// existing destination.
func WritePrivateFileExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, PrivateFileMode)
	if err != nil {
		return err
	}
	written, writeErr := file.Write(data)
	if writeErr == nil && written != len(data) {
		writeErr = io.ErrShortWrite
	}
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return EnsurePrivateFile(path)
}

// WritePrivateFileAtomic replaces a file through a private temporary file in
// the same directory.
func WritePrivateFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if err := temp.Chmod(PrivateFileMode); err != nil {
		_ = temp.Close()
		return err
	}
	written, writeErr := temp.Write(data)
	if writeErr == nil && written != len(data) {
		writeErr = io.ErrShortWrite
	}
	if writeErr == nil {
		writeErr = temp.Sync()
	}
	closeErr := temp.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	return EnsurePrivateFile(path)
}

// Ensure file is private to the user
func EnsurePrivateFile(path string) error {
	logger := corelog.For("config")
	logger.Debug("ensure private file", "path", path, "mode", "0600")
	if err := os.Chmod(path, PrivateFileMode); err != nil {
		logger.Error("chmod private file failed", "path", path, "err", err)
		return err
	}
	logger.Debug("private file ready", "path", path)
	return nil
}
