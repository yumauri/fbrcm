package firebase

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	AuthenticationCredentialsInvalid = "credentials_invalid"
	AuthenticationSetupRequired      = "setup_required"
	AuthenticationRequestFailed      = "request_failed"
)

// AuthenticationError classifies credential loading and identity-provider
// failures without exposing OAuth implementation details to CLI callers.
type AuthenticationError struct {
	Kind       string
	AuthType   string
	Operation  string
	HTTPStatus int
	RemoteCode string
	RetryAfter time.Duration
	Retryable  bool
	Err        error
}

func (e *AuthenticationError) Error() string { return e.Err.Error() }
func (e *AuthenticationError) Unwrap() error { return e.Err }

func credentialAuthenticationError(authType, operation string, err error) error {
	if err == nil {
		return nil
	}
	return &AuthenticationError{Kind: AuthenticationCredentialsInvalid, AuthType: authType, Operation: operation, Err: err}
}

func setupAuthenticationError(authType, operation string, err error) error {
	if err == nil {
		return nil
	}
	return &AuthenticationError{Kind: AuthenticationSetupRequired, AuthType: authType, Operation: operation, Err: err}
}

func authenticationRequestError(authType, operation string, err error) error {
	if err == nil {
		return nil
	}
	var existing *AuthenticationError
	if errors.As(err, &existing) {
		return err
	}

	result := &AuthenticationError{Kind: AuthenticationRequestFailed, AuthType: authType, Operation: operation, Err: err}
	var retrieve *oauth2.RetrieveError
	if errors.As(err, &retrieve) {
		result.RemoteCode = strings.TrimSpace(retrieve.ErrorCode)
		if retrieve.Response != nil {
			result.HTTPStatus = retrieve.Response.StatusCode
			if value := strings.TrimSpace(retrieve.Response.Header.Get("Retry-After")); value != "" {
				if seconds, parseErr := strconv.Atoi(value); parseErr == nil && seconds >= 0 {
					result.RetryAfter = time.Duration(seconds) * time.Second
				}
			}
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		result.Retryable = true
	}
	if result.HTTPStatus == http.StatusRequestTimeout || result.HTTPStatus == http.StatusTooManyRequests || result.HTTPStatus >= 500 {
		result.Retryable = true
	}
	remoteRetryable := strings.EqualFold(result.RemoteCode, "temporarily_unavailable") || strings.EqualFold(result.RemoteCode, "server_error") || strings.EqualFold(result.RemoteCode, "slow_down")
	if remoteRetryable {
		result.Retryable = true
	}
	if result.HTTPStatus >= 400 && result.HTTPStatus < 500 && result.HTTPStatus != http.StatusRequestTimeout && result.HTTPStatus != http.StatusTooManyRequests && !remoteRetryable {
		result.Kind = AuthenticationCredentialsInvalid
	}
	return result
}

func oauthRefreshRequiresAuthorization(err error, hasRefreshToken bool) bool {
	if !hasRefreshToken {
		return true
	}
	var retrieve *oauth2.RetrieveError
	if !errors.As(err, &retrieve) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(retrieve.ErrorCode)) {
	case "invalid_grant", "invalid_token":
		return true
	default:
		return false
	}
}

type authenticationTokenSource struct {
	base      oauth2.TokenSource
	authType  string
	operation string
}

func (s authenticationTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err != nil {
		return nil, authenticationRequestError(s.authType, s.operation, err)
	}
	return token, nil
}
