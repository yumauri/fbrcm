package core

import "github.com/yumauri/fbrcm/core/config"

type Project = config.Project

// AuthPaths describes files owned or used by an auth identity.
type AuthPaths struct {
	AuthConfigPath     string
	ProfileConfigPath  string
	ClientSecretPath   string
	TokenPath          string
	ServiceAccountPath string
}
