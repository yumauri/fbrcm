package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yumauri/fbrcm/core/config"
)

// ListAuth lists configured auth identities.
func (s *Core) ListAuth() ([]config.AuthEntry, string, error) {
	auth, err := config.LoadAuthOrEmpty()
	if err != nil {
		return nil, "", err
	}
	return append([]config.AuthEntry(nil), auth.Auth...), auth.DefaultAuthID, nil
}

// AddOAuthAuth adds or replaces OAuth auth identity.
func (s *Core) AddOAuthAuth(authID, label string, secret []byte) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, err
	}
	previousAuth, hadPrevious := authFile.FindAuth(authID)
	entry := config.DefaultOAuthAuthEntry(authID, label)
	authFile = config.UpsertAuthEntry(authFile, entry)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, err
	}
	secretPath := config.OAuthClientSecretPath(entry)
	tokenPath := config.OAuthTokenPath(entry)
	previousSecret, readErr := os.ReadFile(secretPath)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return config.AuthEntry{}, fmt.Errorf("read existing client secret: %w", readErr)
	}
	secretChanged := readErr == nil && string(previousSecret) != string(secret)
	if err := config.EnsurePrivateDir(filepath.Dir(secretPath)); err != nil {
		return config.AuthEntry{}, fmt.Errorf("create auth dir: %w", err)
	}
	if err := config.EnsurePrivateDir(filepath.Dir(tokenPath)); err != nil {
		return config.AuthEntry{}, fmt.Errorf("create auth cache dir: %w", err)
	}
	if err := config.WritePrivateFile(secretPath, secret); err != nil {
		return config.AuthEntry{}, fmt.Errorf("write client secret: %w", err)
	}
	if secretChanged {
		if err := os.Remove(tokenPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return config.AuthEntry{}, fmt.Errorf("remove token for previous client secret: %w", err)
		}
	}
	if hadPrevious && previousAuth.Type != config.AuthTypeOAuth {
		if err := removeAuthFiles(previousAuth); err != nil {
			return config.AuthEntry{}, err
		}
	}
	s.dropFirebaseService(authID)
	return entry, nil
}

// AddServiceAccountAuth adds or replaces service account auth identity.
func (s *Core) AddServiceAccountAuth(authID, label string, key []byte) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, err
	}
	previous, hadPrevious := authFile.FindAuth(authID)
	entry := config.DefaultServiceAccountAuthEntry(authID, label)
	authFile = config.UpsertAuthEntry(authFile, entry)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, err
	}
	keyPath := config.ServiceAccountKeyPath(entry)
	if err := config.EnsurePrivateDir(filepath.Dir(keyPath)); err != nil {
		return config.AuthEntry{}, fmt.Errorf("create auth dir: %w", err)
	}
	if err := config.WritePrivateFile(keyPath, key); err != nil {
		return config.AuthEntry{}, fmt.Errorf("write service account key: %w", err)
	}
	if hadPrevious && previous.Type != config.AuthTypeServiceAccount {
		if err := removeAuthFiles(previous); err != nil {
			return config.AuthEntry{}, err
		}
	}
	s.dropFirebaseService(authID)
	return entry, nil
}

// AddGCloudAuth adds or replaces gcloud ADC auth identity.
func (s *Core) AddGCloudAuth(authID, label string) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, err
	}
	previous, hadPrevious := authFile.FindAuth(authID)
	entry := config.DefaultGCloudAuthEntry(authID, label)
	authFile = config.UpsertAuthEntry(authFile, entry)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, err
	}
	if hadPrevious {
		if err := removeAuthFiles(previous); err != nil {
			return config.AuthEntry{}, err
		}
	}
	s.dropFirebaseService(authID)
	return entry, nil
}

// AuthPaths gets resolved paths for auth id.
func (s *Core) AuthPaths(authID string) (config.AuthEntry, AuthPaths, error) {
	auth, err := s.authEntry(authID)
	if err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	paths := AuthPaths{
		AuthConfigPath:    config.GetAuthFilePath(),
		ProfileConfigPath: config.GetConfigDirPath(),
	}
	switch auth.Type {
	case config.AuthTypeOAuth:
		paths.ClientSecretPath = config.OAuthClientSecretPath(auth)
		paths.TokenPath = config.OAuthTokenPath(auth)
	case config.AuthTypeServiceAccount:
		paths.ServiceAccountPath = config.ServiceAccountKeyPath(auth)
	}
	return auth, paths, nil
}

// PurgeAuth removes auth identity files and registry entry.
func (s *Core) PurgeAuth(authID string) (config.AuthEntry, AuthPaths, error) {
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	auth, ok := authFile.FindAuth(authID)
	if !ok {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("auth %q is not configured", authID)
	}
	_, paths, err := s.AuthPaths(authID)
	if err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	authFile, _ = config.RemoveAuth(authFile, authID)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	if err := removeFileIfPresent(paths.ClientSecretPath); err != nil {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("remove client secret: %w", err)
	}
	if err := removeFileIfPresent(paths.TokenPath); err != nil {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("remove token: %w", err)
	}
	if err := removeFileIfPresent(paths.ServiceAccountPath); err != nil {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("remove service account key: %w", err)
	}
	s.dropFirebaseService(authID)
	return auth, paths, nil
}
