package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// paths holds paths state used by the config package.
type paths struct {
	// configRootDir stores config root dir for paths.
	configRootDir string
	// cacheRootDir stores cache root dir for paths.
	cacheRootDir string
	// profile stores profile for paths.
	profile string
	// configDir stores config dir for paths.
	configDir string
	// cacheDir stores cache dir for paths.
	cacheDir string
	// projectsFile stores projects file for paths.
	projectsFile string
	// secretFile stores secret file for paths.
	secretFile string
	// tokenFile stores token file for paths.
	tokenFile string
}

var (
	pathsInstance *paths
	pathsOnce     sync.Once
)

// Get application used paths, resolving them once per process
func getPaths() *paths {
	pathsOnce.Do(func() {
		configRootDir := resolveConfigDir()
		cacheRootDir := resolveCacheDir()
		profile := activeProfileOrDefault()
		configDir := filepath.Join(configRootDir, profile)
		cacheDir := filepath.Join(cacheRootDir, profile)

		projectsFile := filepath.Join(configDir, "projects-config.json")
		secretFile := filepath.Join(configDir, "client_secret.json")
		tokenFile := filepath.Join(cacheDir, "token.json")

		pathsInstance = &paths{
			configRootDir: configRootDir,
			cacheRootDir:  cacheRootDir,
			profile:       profile,
			configDir:     configDir,
			cacheDir:      cacheDir,
			projectsFile:  projectsFile,
			secretFile:    secretFile,
			tokenFile:     tokenFile,
		}

		corelog.For("config").Debug("resolved application paths", "config_root_dir", configRootDir, "cache_root_dir", cacheRootDir, "profile", profile, "config_dir", configDir, "cache_dir", cacheDir, "projects_file", projectsFile, "secret_file", secretFile, "token_file", tokenFile)
	})

	return pathsInstance
}

// Reset cached path resolution after profile changes.
func resetPaths() {
	pathsInstance = nil
	pathsOnce = sync.Once{}
}

// Get the path to the config root directory
func GetConfigRootDirPath() string {
	return resolveConfigDir()
}

// Get the path to the cache root directory
func GetCacheRootDirPath() string {
	return resolveCacheDir()
}

// Get the path to the config directory
func GetConfigDirPath() string {
	return getPaths().configDir
}

// Get the path to the cache directory
func GetCacheDirPath() string {
	return getPaths().cacheDir
}

// Get the path to the saved projects list file
func GetProjectsFilePath() string {
	return getPaths().projectsFile
}

// Get the path to the Firebase OAuth secret file
func GetSecretFilePath() string {
	return getPaths().secretFile
}

// Get the path to the Firebase OAuth token file
func GetTokenFilePath() string {
	return getPaths().tokenFile
}

// Resolve location of the config directory, depending on the environment
func resolveConfigDir() string {
	logger := corelog.For("config")
	if override, ok := env.LookupTrimmed(env.ConfigDir); ok {
		logger.Debug("resolved config dir from env override", "env", env.ConfigDir, "path", override)
		return override
	}

	if xdg, ok := env.LookupTrimmed(env.XDGConfigHome); ok {
		path := filepath.Join(xdg, "fbrcm")
		logger.Debug("resolved config dir from xdg config home", "path", path)
		return path
	}

	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".config", "fbrcm")
		logger.Debug("resolved config dir from user home", "path", path)
		return path
	}

	path := filepath.Join(".config", "fbrcm")
	logger.Warn("resolved config dir from relative fallback", "path", path)
	return path
}

// Resolve location of the cache directory, depending on the environment
func resolveCacheDir() string {
	logger := corelog.For("config")
	if override, ok := env.LookupTrimmed(env.CacheDir); ok {
		logger.Debug("resolved cache dir from env override", "env", env.CacheDir, "path", override)
		return override
	}

	if cacheRoot, err := os.UserCacheDir(); err == nil && cacheRoot != "" {
		path := filepath.Join(cacheRoot, "fbrcm")
		logger.Debug("resolved cache dir from user cache dir", "path", path)
		return path
	}

	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".cache", "fbrcm")
		logger.Debug("resolved cache dir from user home", "path", path)
		return path
	}

	logger.Warn("resolved cache dir via config dir fallback")
	return resolveConfigDir()
}
