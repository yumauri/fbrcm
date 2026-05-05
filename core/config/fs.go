package config

import (
	"os"

	corelog "fbrcm/core/log"
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
