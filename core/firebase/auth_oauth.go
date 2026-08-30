package firebase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/yumauri/fbrcm/core/browser"
	"github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type oauthTerminalOutputKey struct{}
type oauthAuthorizationObserverKey struct{}
type oauthInteractionAllowedKey struct{}

// ErrOAuthInteractionRequired reports that usable OAuth credentials are not
// available without starting the desktop authorization flow.
var ErrOAuthInteractionRequired = errors.New("OAuth browser authorization is required")

// OAuthAuthorizationEvent reports the interactive portion of a desktop OAuth
// flow. A start event contains URL; a completion event has Done set.
type OAuthAuthorizationEvent struct {
	URL    string
	Done   bool
	Err    error
	Cancel context.CancelFunc
}

// WithOAuthAuthorizationObserver reports desktop OAuth progress without
// coupling the Firebase layer to a particular user interface.
func WithOAuthAuthorizationObserver(ctx context.Context, observer func(OAuthAuthorizationEvent)) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, oauthAuthorizationObserverKey{}, observer)
}

func notifyOAuthAuthorization(ctx context.Context, event OAuthAuthorizationEvent) {
	if ctx == nil {
		return
	}
	observer, _ := ctx.Value(oauthAuthorizationObserverKey{}).(func(OAuthAuthorizationEvent))
	if observer != nil {
		observer(event)
	}
}

// WithOAuthTerminalOutput controls whether the desktop OAuth flow writes its
// authorization URL and status to stderr. It defaults to enabled for CLI
// callers; full-screen TUI callers disable it to avoid corrupting rendering.
func WithOAuthTerminalOutput(ctx context.Context, enabled bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, oauthTerminalOutputKey{}, enabled)
}

// WithOAuthInteractionAllowed controls whether a command may start the desktop
// browser authorization flow. It defaults to true for human CLI and TUI use.
func WithOAuthInteractionAllowed(ctx context.Context, allowed bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, oauthInteractionAllowedKey{}, allowed)
}

func oauthInteractionAllowed(ctx context.Context) bool {
	if ctx == nil {
		return true
	}
	allowed, configured := ctx.Value(oauthInteractionAllowedKey{}).(bool)
	return !configured || allowed
}

func oauthTerminalOutputEnabled(ctx context.Context) bool {
	if ctx == nil {
		return true
	}
	enabled, configured := ctx.Value(oauthTerminalOutputKey{}).(bool)
	return !configured || enabled
}

// Create HTTP client configured with OAuth2 credentials
func oauthHTTPClient(ctx context.Context, clientSecretPath, tokenPath string, autoOpen bool) (*http.Client, error) {
	logger := corelog.For("firebase")
	logger.Info("load oauth client secret", "path", clientSecretPath)

	clientSecretData, err := os.ReadFile(clientSecretPath)
	if err != nil {
		logger.Error("read oauth client secret failed", "path", clientSecretPath, "err", err)
		return nil, fmt.Errorf("reading OAuth client secret: %w", err)
	}

	oauthCfg, err := google.ConfigFromJSON(clientSecretData, cloudPlatformScope)
	if err != nil {
		logger.Error("parse oauth client secret failed", "path", clientSecretPath, "err", err)
		return nil, credentialAuthenticationError(config.AuthTypeOAuth, "load_credentials", fmt.Errorf("parsing OAuth client secret: %w", err))
	}
	return oauthHTTPClientWithConfig(ctx, config.AuthTypeOAuth, oauthCfg, tokenPath, autoOpen)
}

func googleOAuthHTTPClient(ctx context.Context, credentials OAuthClientCredentials, tokenPath string, autoOpen bool) (*http.Client, error) {
	oauthCfg, err := credentials.config()
	if err != nil {
		return nil, credentialAuthenticationError(config.AuthTypeGoogle, "load_credentials", err)
	}
	return oauthHTTPClientWithConfig(ctx, config.AuthTypeGoogle, oauthCfg, tokenPath, autoOpen)
}

func oauthHTTPClientWithConfig(ctx context.Context, authType string, oauthCfg *oauth2.Config, tokenPath string, autoOpen bool) (*http.Client, error) {
	logger := corelog.For("firebase")
	persistAuthState := !IsDryRun(ctx)

	tok, err := readCachedToken(tokenPath)
	if err != nil {
		logger.Error("read cached oauth token failed", "err", err)
		return nil, err
	}
	if tok == nil {
		if !oauthInteractionAllowed(ctx) {
			logger.Info("oauth token cache miss requires interaction in non-interactive context")
			return nil, ErrOAuthInteractionRequired
		}
		if IsOffline() {
			logger.Warn("offline mode, cannot start oauth authorization flow")
			return nil, ErrOffline
		}
		logger.Warn("oauth token cache miss; starting authorization flow")
		tok, err = authorizeDesktopClient(ctx, oauthCfg, true, autoOpen)
		if err != nil {
			return nil, err
		}
		if persistAuthState {
			if err := writeCachedToken(tokenPath, tok); err != nil {
				return nil, err
			}
		} else {
			logger.Warn("dry run, skip initial oauth token cache save")
		}
	}

	if IsOffline() {
		logger.Warn("offline mode, using cached oauth token without refresh", "has_refresh_token", tok.RefreshToken != "")
		client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tok))
		return wrapAuthHTTPClient(ctx, client), nil
	}

	baseTokenSource := oauthCfg.TokenSource(ctx, tok)
	tokenSource := &persistingTokenSource{
		base:     baseTokenSource,
		authType: authType,
		persist:  persistAuthState,
		path:     tokenPath,
	}

	hadRefreshToken, err := refreshOAuthToken(tokenSource, tok)
	if err != nil {
		if !oauthRefreshRequiresAuthorization(err, hadRefreshToken) {
			return nil, authenticationRequestError(authType, "refresh_token", err)
		}
		logger.Warn("oauth token refresh requires reauthorization", "has_refresh_token", hadRefreshToken)
		tok, err = authorizeDesktopClient(ctx, oauthCfg, true, autoOpen)
		if err != nil {
			return nil, err
		}
		if persistAuthState {
			if err := writeCachedToken(tokenPath, tok); err != nil {
				return nil, err
			}
		} else {
			logger.Warn("dry run, skip oauth token cache save after reauthorization")
		}
		baseTokenSource = oauthCfg.TokenSource(ctx, tok)
		tokenSource = &persistingTokenSource{
			base:     baseTokenSource,
			authType: authType,
			persist:  persistAuthState,
			path:     tokenPath,
		}
	}

	logger.Debug("oauth http client ready")
	rotatingSource := newRotatingOAuthTokenSource(tokenSource, tok)
	client, err := newRecoveringOAuthClient(
		ctx,
		rotatingSource,
		func(cached *oauth2.Token) (oauth2.TokenSource, *oauth2.Token, error) {
			return recoverRejectedOAuthToken(
				ctx,
				authType,
				oauthCfg,
				cached,
				tokenPath,
				persistAuthState,
				autoOpen,
			)
		},
	)
	if err != nil {
		return nil, err
	}
	return wrapAuthHTTPClient(ctx, client), nil
}

func refreshOAuthToken(source oauth2.TokenSource, cached *oauth2.Token) (bool, error) {
	_, err := source.Token()
	if err != nil {
		return cached != nil && cached.RefreshToken != "", err
	}
	return false, nil
}

// Authorizes a desktop client using OAuth2 and returns the OAuth token
func authorizeDesktopClient(ctx context.Context, oauthCfg *oauth2.Config, forceConsent bool, autoOpen bool) (token *oauth2.Token, returnErr error) {
	logger := corelog.For("firebase")
	logger.Info("start oauth desktop authorization", "force_consent", forceConsent, "auto_open", autoOpen)
	if !oauthInteractionAllowed(ctx) {
		logger.Info("oauth desktop authorization blocked by non-interactive context")
		return nil, ErrOAuthInteractionRequired
	}
	oauthCfgCopy := *oauthCfg
	oauthCfg = &oauthCfgCopy
	authorizationCtx, cancelAuthorization := context.WithCancel(ctx)
	defer cancelAuthorization()

	state, err := randomToken(32)
	if err != nil {
		logger.Error("generate oauth state failed", "err", err)
		return nil, fmt.Errorf("generating OAuth state: %w", err)
	}
	verifier, err := randomToken(32)
	if err != nil {
		logger.Error("generate pkce verifier failed", "err", err)
		return nil, fmt.Errorf("generating PKCE verifier: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv, redirectURL, err := startLoopbackServer(state, codeCh, errCh)
	if err != nil {
		return nil, err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	oauthCfg.RedirectURL = redirectURL
	authCodeOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	if forceConsent {
		authCodeOpts = append(authCodeOpts, oauth2.ApprovalForce)
	}
	authURL := oauthCfg.AuthCodeURL(state, authCodeOpts...)
	logger.Info("oauth authorization url ready", "url", redactedURLStringValue(authURL))
	notifyOAuthAuthorization(ctx, OAuthAuthorizationEvent{URL: authURL, Cancel: cancelAuthorization})
	defer func() {
		notifyOAuthAuthorization(ctx, OAuthAuthorizationEvent{Done: true, Err: returnErr})
	}()

	terminalOutput := oauthTerminalOutputEnabled(ctx)
	terminalWriter := corelog.TerminalOutput()
	if terminalOutput {
		_, _ = fmt.Fprintln(terminalWriter, "Open this URL in your browser to authorize fbrcm:")
		_, _ = fmt.Fprintln(terminalWriter, authURL)
	}
	if autoOpen {
		if err := browser.OpenURL(authURL); err != nil {
			logger.Warn("open browser automatically failed", "err", err)
			if terminalOutput {
				_, _ = fmt.Fprintf(terminalWriter, "Could not open browser automatically: %v\n", err)
			}
		}
	} else {
		logger.Info("browser auto-open disabled")
	}
	if terminalOutput {
		_, _ = fmt.Fprintln(terminalWriter, "Waiting for OAuth callback on local loopback server...")
	}
	logger.Info("waiting for oauth callback")

	timer := time.NewTimer(2 * time.Minute)
	defer timer.Stop()
	select {
	case code := <-codeCh:
		logger.Info("oauth callback received; exchanging code")
		tok, err := oauthCfg.Exchange(
			authorizationCtx,
			code,
			oauth2.SetAuthURLParam("code_verifier", verifier),
		)
		if err != nil {
			logger.Error("oauth code exchange failed", "err", err)
			return nil, fmt.Errorf("exchanging OAuth code: %w", err)
		}
		logger.Info("oauth authorization complete")
		return tok, nil
	case err := <-errCh:
		logger.Error("oauth callback failed", "err", err)
		return nil, err
	case <-authorizationCtx.Done():
		logger.Info("oauth authorization canceled", "err", authorizationCtx.Err())
		return nil, fmt.Errorf("OAuth authorization canceled: %w", authorizationCtx.Err())
	case <-timer.C:
		logger.Error("oauth callback timed out")
		return nil, fmt.Errorf("timed out waiting for OAuth callback")
	}
}

// startLoopbackServer starts a loopback OAuth callback server and redacts callback URLs in logs.
func startLoopbackServer(expectedState string, codeCh chan<- string, errCh chan<- error) (*http.Server, string, error) {
	logger := corelog.For("firebase")
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: mux,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		logger.Error("start oauth callback listener failed", "addr", srv.Addr, "err", err)
		return nil, "", fmt.Errorf("starting local OAuth callback listener: %w", err)
	}

	redirectURL := fmt.Sprintf("http://%s/oauth2callback", ln.Addr().String())
	logger.Info("oauth callback listener started", "addr", ln.Addr().String())
	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("http request", "method", r.Method, "url", redactedURLString(r.URL), "headers", formatHeaders(r.Header))

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			logger.Error("oauth callback returned error", "err", errMsg)
			http.Error(w, "OAuth authorization failed", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("OAuth authorization failed: %s", errMsg):
			default:
			}
			return
		}

		if r.URL.Query().Get("state") != expectedState {
			logger.Error("oauth callback state mismatch")
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("invalid OAuth state in callback"):
			default:
			}
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			logger.Error("oauth callback missing code")
			http.Error(w, "Missing OAuth code", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("missing OAuth code in callback"):
			default:
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "Authorization complete. Return to fbrcm.")
		safeURL := redactedURLString(r.URL)
		logger.Info("http response", "method", r.Method, "url", safeURL, "status", "200 OK")
		logger.Debug("http response headers", "method", r.Method, "url", safeURL, "status", http.StatusText(http.StatusOK), "headers", formatHeaders(w.Header()))
		select {
		case codeCh <- code:
			logger.Info("oauth callback accepted")
		default:
		}
	})

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Error("oauth callback server failed", "err", err)
			select {
			case errCh <- fmt.Errorf("OAuth callback server failed: %w", err):
			default:
			}
		}
	}()

	return srv, redirectURL, nil
}

// Generates a random token of the given size using crypto/rand
func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// Generates a PKCE challenge for the given verifier
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
