package config

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/yumauri/fbrcm/core/strfold"
)

const (
	AuthConfigVersion      = 1
	AuthTypeOAuth          = "oauth"
	AuthTypeServiceAccount = "service-account"
	AuthTypeGCloud         = "gcloud"
)

// AuthFile holds auth registry state.
type AuthFile struct {
	Version       int         `json:"version"`
	DefaultAuthID string      `json:"default_auth_id"`
	Auth          []AuthEntry `json:"auth"`
}

// AuthEntry describes one configured auth identity.
type AuthEntry struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Label              string `json:"label"`
	ClientSecretPath   string `json:"client_secret_path,omitempty"`
	TokenPath          string `json:"token_path,omitempty"`
	ServiceAccountPath string `json:"service_account_path,omitempty"`
}

// ValidateAuthID validates auth id as a single safe path segment.
func ValidateAuthID(id string) error {
	return validatePathSegment(id, "auth id")
}

// LoadAuth loads auth registry.
func LoadAuth() (*AuthFile, error) {
	path := GetAuthFilePath()
	var file AuthFile
	if err := readJSONFile(path, &file); err != nil {
		if isDecodeError(err) {
			return nil, fmt.Errorf("decode auth config: %w", err)
		}
		return nil, fmt.Errorf("read auth config: %w", err)
	}
	if err := validateAuthFile(&file); err != nil {
		return nil, err
	}
	return &file, nil
}

// SaveAuth saves auth registry.
func SaveAuth(file *AuthFile) error {
	if file == nil {
		return fmt.Errorf("auth config is nil")
	}
	file.Version = AuthConfigVersion
	slices.SortStableFunc(file.Auth, func(left, right AuthEntry) int {
		return strfold.Compare(left.ID, right.ID)
	})
	if file.DefaultAuthID == "" && len(file.Auth) > 0 {
		file.DefaultAuthID = file.Auth[0].ID
	}
	if err := validateAuthFile(file); err != nil {
		return err
	}
	if err := EnsurePrivateDir(GetConfigDirPath()); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	path := GetAuthFilePath()
	if err := writeJSONFile(path, file); err != nil {
		if isEncodeError(err) {
			return fmt.Errorf("encode auth config: %w", err)
		}
		return fmt.Errorf("write auth config: %w", err)
	}
	return nil
}

// DefaultOAuthAuthEntry builds default OAuth auth entry for id.
func DefaultOAuthAuthEntry(id, label string) AuthEntry {
	return AuthEntry{
		ID:               id,
		Type:             AuthTypeOAuth,
		Label:            label,
		ClientSecretPath: filepath.ToSlash(filepath.Join("auth", id, "client-secret.json")),
		TokenPath:        filepath.ToSlash(filepath.Join("auth", id, "token.json")),
	}
}

// DefaultServiceAccountAuthEntry builds default service account auth entry for id.
func DefaultServiceAccountAuthEntry(id, label string) AuthEntry {
	return AuthEntry{
		ID:                 id,
		Type:               AuthTypeServiceAccount,
		Label:              label,
		ServiceAccountPath: filepath.ToSlash(filepath.Join("auth", id, "service-account.json")),
	}
}

// DefaultGCloudAuthEntry builds default gcloud ADC auth entry for id.
func DefaultGCloudAuthEntry(id, label string) AuthEntry {
	return AuthEntry{
		ID:    id,
		Type:  AuthTypeGCloud,
		Label: label,
	}
}

// UpsertAuthEntry upserts auth entry and makes first entry default.
func UpsertAuthEntry(file *AuthFile, entry AuthEntry) *AuthFile {
	if file == nil {
		file = &AuthFile{Version: AuthConfigVersion}
	}
	replaced := false
	for i := range file.Auth {
		if file.Auth[i].ID == entry.ID {
			file.Auth[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		file.Auth = append(file.Auth, entry)
	}
	if file.DefaultAuthID == "" {
		file.DefaultAuthID = entry.ID
	}
	return file
}

// FindAuth finds an auth entry by id.
func (f *AuthFile) FindAuth(id string) (AuthEntry, bool) {
	if f == nil {
		return AuthEntry{}, false
	}
	for _, entry := range f.Auth {
		if entry.ID == id {
			return entry, true
		}
	}
	return AuthEntry{}, false
}

// OAuthClientSecretPath resolves OAuth client secret path for entry.
func OAuthClientSecretPath(entry AuthEntry) string {
	return resolveConfigAuthPath(entry.ClientSecretPath)
}

// OAuthTokenPath resolves OAuth token path for entry.
func OAuthTokenPath(entry AuthEntry) string {
	return resolveCacheAuthPath(entry.TokenPath)
}

// ServiceAccountKeyPath resolves service account key path for entry.
func ServiceAccountKeyPath(entry AuthEntry) string {
	return resolveConfigAuthPath(entry.ServiceAccountPath)
}

// RemoveAuth removes auth entry by id.
func RemoveAuth(file *AuthFile, id string) (*AuthFile, bool) {
	if file == nil {
		return &AuthFile{Version: AuthConfigVersion}, false
	}
	next := make([]AuthEntry, 0, len(file.Auth))
	removed := false
	for _, entry := range file.Auth {
		if entry.ID == id {
			removed = true
			continue
		}
		next = append(next, entry)
	}
	file.Auth = next
	if file.DefaultAuthID == id {
		file.DefaultAuthID = ""
		if len(file.Auth) > 0 {
			file.DefaultAuthID = file.Auth[0].ID
		}
	}
	return file, removed
}

func validateAuthFile(file *AuthFile) error {
	if file.Version != AuthConfigVersion {
		return fmt.Errorf("unsupported auth config version %d", file.Version)
	}
	seen := map[string]struct{}{}
	defaultSeen := file.DefaultAuthID == ""
	for _, entry := range file.Auth {
		if err := ValidateAuthID(entry.ID); err != nil {
			return err
		}
		if _, ok := seen[entry.ID]; ok {
			return fmt.Errorf("duplicate auth id %q", entry.ID)
		}
		seen[entry.ID] = struct{}{}
		switch entry.Type {
		case AuthTypeOAuth:
			if strings.TrimSpace(entry.ClientSecretPath) == "" {
				return fmt.Errorf("auth %s client secret path is empty", entry.ID)
			}
			if err := validateAuthStoragePath(entry.ClientSecretPath, entry.ID, "client-secret.json"); err != nil {
				return fmt.Errorf("auth %s client secret path: %w", entry.ID, err)
			}
			if strings.TrimSpace(entry.TokenPath) == "" {
				return fmt.Errorf("auth %s token path is empty", entry.ID)
			}
			if err := validateAuthStoragePath(entry.TokenPath, entry.ID, "token.json"); err != nil {
				return fmt.Errorf("auth %s token path: %w", entry.ID, err)
			}
		case AuthTypeServiceAccount:
			if strings.TrimSpace(entry.ServiceAccountPath) == "" {
				return fmt.Errorf("auth %s service account path is empty", entry.ID)
			}
			if err := validateAuthStoragePath(entry.ServiceAccountPath, entry.ID, "service-account.json"); err != nil {
				return fmt.Errorf("auth %s service account path: %w", entry.ID, err)
			}
		case AuthTypeGCloud:
		default:
			return fmt.Errorf("unsupported auth type %q for %s", entry.Type, entry.ID)
		}
		if entry.ID == file.DefaultAuthID {
			defaultSeen = true
		}
	}
	if !defaultSeen {
		return fmt.Errorf("default auth %q is not configured", file.DefaultAuthID)
	}
	return nil
}

func validateAuthStoragePath(path, authID, filename string) error {
	if filepath.IsAbs(path) {
		return fmt.Errorf("must be relative")
	}
	if strings.Contains(path, `\`) {
		return fmt.Errorf("cannot contain path separators other than slash")
	}

	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean != path {
		return fmt.Errorf("must be clean")
	}

	expected := filepath.ToSlash(filepath.Join("auth", authID, filename))
	if clean != expected {
		return fmt.Errorf("must be %q", expected)
	}
	return nil
}

func resolveConfigAuthPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(GetConfigDirPath(), filepath.FromSlash(path))
}

func resolveCacheAuthPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(GetCacheDirPath(), filepath.FromSlash(path))
}

// LoadAuthOrEmpty loads auth registry or returns an empty one on cache miss.
func LoadAuthOrEmpty() (*AuthFile, error) {
	file, err := LoadAuth()
	if err != nil {
		if isNotExist(err) {
			return &AuthFile{Version: AuthConfigVersion}, nil
		}
		return nil, err
	}
	return file, nil
}
