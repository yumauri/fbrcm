package firebase

import (
	"errors"
	"testing"

	"golang.org/x/oauth2"
)

type failingOAuthTokenSource struct {
	err error
}

func (s failingOAuthTokenSource) Token() (*oauth2.Token, error) {
	return nil, s.err
}

func TestRefreshOAuthTokenRetainsCachedRefreshStateOnNilErrorResult(t *testing.T) {
	refreshErr := errors.New("refresh failed")
	cached := &oauth2.Token{AccessToken: "expired", RefreshToken: "refresh"}

	hadRefreshToken, err := refreshOAuthToken(failingOAuthTokenSource{err: refreshErr}, cached)

	if !errors.Is(err, refreshErr) {
		t.Fatalf("refreshOAuthToken error = %v, want %v", err, refreshErr)
	}
	if !hadRefreshToken {
		t.Fatal("refreshOAuthToken lost cached refresh-token state")
	}
}

func TestRefreshOAuthTokenHandlesMissingCachedToken(t *testing.T) {
	refreshErr := errors.New("refresh failed")

	hadRefreshToken, err := refreshOAuthToken(failingOAuthTokenSource{err: refreshErr}, nil)

	if !errors.Is(err, refreshErr) || hadRefreshToken {
		t.Fatalf("refreshOAuthToken = %t, %v; want false, refresh error", hadRefreshToken, err)
	}
}
