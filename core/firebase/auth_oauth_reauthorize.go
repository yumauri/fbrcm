package firebase

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"

	corelog "github.com/yumauri/fbrcm/core/log"
)

type rotatingOAuthTokenSource struct {
	mu         sync.Mutex
	source     oauth2.TokenSource
	current    *oauth2.Token
	generation uint64
}

func newRotatingOAuthTokenSource(source oauth2.TokenSource, current *oauth2.Token) *rotatingOAuthTokenSource {
	return &rotatingOAuthTokenSource{source: source, current: cloneOAuthToken(current)}
}

func (s *rotatingOAuthTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, err := s.source.Token()
	if err == nil {
		s.current = cloneOAuthToken(token)
	}
	return token, err
}

func (s *rotatingOAuthTokenSource) snapshot() (uint64, *oauth2.Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.generation, cloneOAuthToken(s.current)
}

func (s *rotatingOAuthTokenSource) replace(source oauth2.TokenSource, current *oauth2.Token) {
	s.mu.Lock()
	s.source = source
	s.current = cloneOAuthToken(current)
	s.generation++
	s.mu.Unlock()
}

func cloneOAuthToken(token *oauth2.Token) *oauth2.Token {
	if token == nil {
		return nil
	}
	cloned := *token
	return &cloned
}

type oauthTokenRecovery func(*oauth2.Token) (oauth2.TokenSource, *oauth2.Token, error)

func newRecoveringOAuthClient(
	ctx context.Context,
	tokens *rotatingOAuthTokenSource,
	recoverToken oauthTokenRecovery,
) (*http.Client, error) {
	client := oauth2.NewClient(ctx, tokens)
	oauthTransport, ok := client.Transport.(*oauth2.Transport)
	if !ok {
		return nil, fmt.Errorf("create OAuth client: unexpected transport %T", client.Transport)
	}
	// oauth2.NewClient adds its own ReuseTokenSource around tokens. That cache
	// cannot observe replacements made after a 401, so use the rotating source
	// directly as the single source of token caching and replacement.
	oauthTransport.Source = tokens
	client.Transport = newOAuthUnauthorizedTransport(oauthTransport, tokens, recoverToken)
	return client, nil
}

type oauthUnauthorizedTransport struct {
	base          http.RoundTripper
	tokens        *rotatingOAuthTokenSource
	recoverToken  oauthTokenRecovery
	recoverTokenM sync.Mutex
}

func newOAuthUnauthorizedTransport(
	base http.RoundTripper,
	tokens *rotatingOAuthTokenSource,
	recoverToken oauthTokenRecovery,
) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &oauthUnauthorizedTransport{base: base, tokens: tokens, recoverToken: recoverToken}
}

func (t *oauthUnauthorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	generation, _ := t.tokens.snapshot()
	response, err := t.base.RoundTrip(req)
	if err != nil || response == nil || response.StatusCode != http.StatusUnauthorized || !requestCanRetry(req) {
		return response, err
	}

	closeRetryResponse(response)
	if err := t.recoverAfterUnauthorized(generation); err != nil {
		return nil, err
	}
	retry, err := cloneRequest(req, 2)
	if err != nil {
		return nil, err
	}
	return t.base.RoundTrip(retry)
}

func (t *oauthUnauthorizedTransport) recoverAfterUnauthorized(observedGeneration uint64) error {
	t.recoverTokenM.Lock()
	defer t.recoverTokenM.Unlock()

	generation, cached := t.tokens.snapshot()
	if generation != observedGeneration {
		return nil
	}
	source, token, err := t.recoverToken(cached)
	if err != nil {
		return err
	}
	t.tokens.replace(source, token)
	return nil
}

func recoverRejectedOAuthToken(
	ctx context.Context,
	authType string,
	oauthCfg *oauth2.Config,
	cached *oauth2.Token,
	tokenPath string,
	persist bool,
	autoOpen bool,
) (oauth2.TokenSource, *oauth2.Token, error) {
	logger := corelog.For("firebase")
	logger.Warn("oauth access token rejected; attempting forced refresh")

	token, refreshErr := forceRefreshOAuthToken(ctx, authType, oauthCfg, cached)
	if refreshErr != nil {
		if !oauthRefreshRequiresAuthorization(refreshErr, cached != nil && cached.RefreshToken != "") {
			return nil, nil, authenticationRequestError(authType, "refresh_token", refreshErr)
		}
		logger.Warn("forced oauth token refresh requires reauthorization", "err", refreshErr)
		if IsOffline() {
			return nil, nil, ErrOffline
		}
		var err error
		token, err = authorizeDesktopClient(ctx, oauthCfg, true, autoOpen)
		if err != nil {
			return nil, nil, err
		}
	} else {
		logger.Info("recovered rejected oauth access token with refresh token")
	}

	if persist {
		if err := writeCachedToken(tokenPath, token); err != nil {
			return nil, nil, err
		}
	} else {
		logger.Warn("dry run, skip recovered oauth token cache save")
	}
	source := &persistingTokenSource{
		base:      oauthCfg.TokenSource(ctx, token),
		lastToken: cloneOAuthToken(token),
		authType:  authType,
		persist:   persist,
		path:      tokenPath,
	}
	return source, token, nil
}

func forceRefreshOAuthToken(ctx context.Context, authType string, oauthCfg *oauth2.Config, cached *oauth2.Token) (*oauth2.Token, error) {
	if cached == nil || cached.RefreshToken == "" {
		return nil, fmt.Errorf("OAuth refresh token is missing")
	}
	expired := *cached
	expired.AccessToken = ""
	expired.Expiry = time.Unix(0, 0)
	token, err := oauthCfg.TokenSource(ctx, &expired).Token()
	if err != nil {
		return nil, authenticationRequestError(authType, "refresh_token", err)
	}
	return token, nil
}
