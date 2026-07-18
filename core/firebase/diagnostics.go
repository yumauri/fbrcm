package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/yumauri/fbrcm/core/config"
)

// AuthDiagnostic describes locally available credentials without exposing
// credential values.
type AuthDiagnostic struct {
	CredentialPath    string    `json:"credential_path,omitempty"`
	TokenPath         string    `json:"token_path,omitempty"`
	TokenExpiry       time.Time `json:"token_expiry,omitzero"`
	TokenExpired      bool      `json:"token_expired,omitempty"`
	HasRefreshToken   bool      `json:"has_refresh_token,omitempty"`
	TokenError        string    `json:"token_error,omitempty"`
	CredentialWarning string    `json:"credential_warning,omitempty"`
}

// ValidateOAuthClientSecret checks that data is a Google OAuth client
// configuration before it is persisted by an auth-management surface.
func ValidateOAuthClientSecret(data []byte) error {
	if _, err := google.ConfigFromJSON(data, cloudPlatformScope); err != nil {
		return fmt.Errorf("parse OAuth client secret: %w", err)
	}
	return nil
}

// ValidateServiceAccountKey checks that data is a Google service account key
// before it is persisted by an auth-management surface.
func ValidateServiceAccountKey(data []byte) error {
	if _, err := google.JWTConfigFromJSON(data, cloudPlatformScope); err != nil {
		return fmt.Errorf("parse service account key: %w", err)
	}
	return nil
}

// InspectAuth validates local credential and token files without contacting
// Google or starting an authorization flow.
func InspectAuth(auth config.AuthEntry) (AuthDiagnostic, error) {
	switch auth.Type {
	case config.AuthTypeOAuth:
		secretPath := config.OAuthClientSecretPath(auth)
		secret, err := os.ReadFile(secretPath)
		if err != nil {
			return AuthDiagnostic{CredentialPath: secretPath}, fmt.Errorf("read OAuth client secret: %w", err)
		}
		if _, err := google.ConfigFromJSON(secret, cloudPlatformScope); err != nil {
			return AuthDiagnostic{CredentialPath: secretPath}, fmt.Errorf("parse OAuth client secret: %w", err)
		}
		tokenPath := config.OAuthTokenPath(auth)
		token, err := readCachedToken(tokenPath)
		if err != nil {
			return AuthDiagnostic{CredentialPath: secretPath, TokenPath: tokenPath, TokenError: err.Error()}, nil
		}
		if token == nil {
			return AuthDiagnostic{CredentialPath: secretPath, TokenPath: tokenPath, TokenError: "OAuth token is missing"}, nil
		}
		return AuthDiagnostic{
			CredentialPath:  secretPath,
			TokenPath:       tokenPath,
			TokenExpiry:     token.Expiry,
			TokenExpired:    !token.Valid(),
			HasRefreshToken: strings.TrimSpace(token.RefreshToken) != "",
		}, nil
	case config.AuthTypeServiceAccount:
		path := config.ServiceAccountKeyPath(auth)
		key, err := os.ReadFile(path)
		if err != nil {
			return AuthDiagnostic{CredentialPath: path}, fmt.Errorf("read service account key: %w", err)
		}
		if _, err := google.JWTConfigFromJSON(key, cloudPlatformScope); err != nil {
			return AuthDiagnostic{CredentialPath: path}, fmt.Errorf("parse service account key: %w", err)
		}
		return AuthDiagnostic{CredentialPath: path}, nil
	case config.AuthTypeGCloud:
		path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if strings.TrimSpace(path) == "" {
			path = wellKnownADCFile()
		}
		if _, err := os.Stat(path); err != nil {
			return AuthDiagnostic{CredentialPath: path, CredentialWarning: "ADC file not found; checking the default credential chain live"}, nil
		}
		return AuthDiagnostic{CredentialPath: path}, nil
	default:
		return AuthDiagnostic{}, fmt.Errorf("unsupported auth type %q", auth.Type)
	}
}

func diagnosticOAuthHTTPClient(ctx context.Context, clientSecretPath, tokenPath string) (*http.Client, error) {
	secret, err := os.ReadFile(clientSecretPath)
	if err != nil {
		return nil, fmt.Errorf("reading OAuth client secret: %w", err)
	}
	oauthCfg, err := google.ConfigFromJSON(secret, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("parsing OAuth client secret: %w", err)
	}
	token, err := readCachedToken(tokenPath)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, fmt.Errorf("OAuth token is missing; run `fbrcm auth login`")
	}
	if IsOffline() {
		return wrapAuthHTTPClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))), nil
	}
	refreshed, err := oauthCfg.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("refresh OAuth token: %w", err)
	}
	return wrapAuthHTTPClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(refreshed))), nil
}

// TestProjectPermissions returns the requested IAM permissions granted on a
// Firebase project.
func (s *Service) TestProjectPermissions(ctx context.Context, projectID string, permissions []string) ([]string, error) {
	endpoint := fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v1/projects/%s:testIamPermissions", projectID)
	body, err := json.Marshal(map[string][]string{"permissions": permissions})
	if err != nil {
		return nil, fmt.Errorf("encode permission request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create permission request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	s.setQuotaProject(req, projectID)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("test Firebase permissions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read permission response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("permission API returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}
	var payload struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, fmt.Errorf("decode permission response: %w", err)
	}
	return payload.Permissions, nil
}
