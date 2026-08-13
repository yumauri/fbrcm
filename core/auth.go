package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

const authSetupHint = "Set up authentication by running `fbrcm` for guided setup, or see `fbrcm auth add --help` for CLI options."

// AuthError classifies expected authentication registry and selection failures
// without coupling core services to the CLI contract package.
type AuthError struct {
	Kind   string
	AuthID string
	Err    error
}

// OAuthInteractionError identifies an auth identity that must be authorized by
// a human before a non-interactive command can use it.
type OAuthInteractionError struct {
	AuthID string
	Err    error
}

func (e *OAuthInteractionError) Error() string { return e.Err.Error() }
func (e *OAuthInteractionError) Unwrap() error { return e.Err }

func (e *AuthError) Error() string {
	if e.Kind == "setup_required" {
		return e.Err.Error() + "\n\n" + authSetupHint
	}
	return e.Err.Error()
}

func (e *AuthError) Unwrap() error { return e.Err }

func loadAuthWithSetupHint() (*config.AuthFile, error) {
	authFile, err := config.LoadAuth()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &AuthError{Kind: "setup_required", Err: err}
		}
		return nil, &AuthError{Kind: "configuration", Err: err}
	}
	return authFile, nil
}

func loadRequiredAuth() (*config.AuthFile, error) {
	authFile, err := loadAuthWithSetupHint()
	if err != nil {
		return nil, err
	}
	if len(authFile.Auth) == 0 {
		return nil, &AuthError{Kind: "setup_required", Err: errors.New("no auth identities configured")}
	}
	return authFile, nil
}

func authNotConfiguredError(authFile *config.AuthFile, authID string) error {
	err := fmt.Errorf("auth %q is not configured", authID)
	if len(authFile.Auth) == 0 {
		return &AuthError{Kind: "setup_required", AuthID: authID, Err: err}
	}
	return &AuthError{Kind: "not_found", AuthID: authID, Err: err}
}

func loadAuthOrEmpty() (*config.AuthFile, error) {
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return nil, &AuthError{Kind: "configuration", Err: err}
	}
	return authFile, nil
}

// ListAuth lists configured auth identities.
func (s *Core) ListAuth() ([]config.AuthEntry, string, error) {
	auth, err := loadAuthOrEmpty()
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
	if err := firebase.ValidateOAuthClientSecret(secret); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := loadAuthOrEmpty()
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
	if err := firebase.ValidateServiceAccountKey(key); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := loadAuthOrEmpty()
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
	authFile, err := loadAuthOrEmpty()
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

// DeleteAuth removes an auth identity's files and registry entry.
func (s *Core) DeleteAuth(authID string) (config.AuthEntry, AuthPaths, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, AuthPaths{}, &AuthError{Kind: "invalid_argument", AuthID: authID, Err: err}
	}
	authFile, err := loadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	auth, ok := authFile.FindAuth(authID)
	if !ok {
		return config.AuthEntry{}, AuthPaths{}, authNotConfiguredError(authFile, authID)
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
