package firebase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"fbrcm/core/browser"
	"fbrcm/core/config"
	corelog "fbrcm/core/log"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

// Create HTTP client configured with OAuth2 credentials
func oauthHTTPClient(ctx context.Context, autoOpen bool) (*http.Client, error) {
	logger := corelog.For("firebase")
	persistAuthState := !IsDryRun(ctx)
	secretPath := config.GetSecretFilePath()
	logger.Info("load oauth client secret", "path", secretPath)

	clientSecretData, err := os.ReadFile(secretPath)
	if err != nil {
		logger.Error("read oauth client secret failed", "path", secretPath, "err", err)
		return nil, fmt.Errorf("reading OAuth client secret: %w", err)
	}

	oauthCfg, err := google.ConfigFromJSON(clientSecretData, cloudPlatformScope)
	if err != nil {
		logger.Error("parse oauth client secret failed", "path", secretPath, "err", err)
		return nil, fmt.Errorf("parsing OAuth client secret: %w", err)
	}

	tok, err := readCachedToken()
	if err != nil {
		logger.Error("read cached oauth token failed", "err", err)
		return nil, err
	}
	if tok == nil {
		logger.Warn("oauth token cache miss; starting authorization flow")
		tok, err = authorizeDesktopClient(ctx, oauthCfg, true, autoOpen)
		if err != nil {
			return nil, err
		}
		if persistAuthState {
			if err := writeCachedToken(tok); err != nil {
				return nil, err
			}
		} else {
			logger.Warn("dry run, skip initial oauth token cache save")
		}
	}

	baseTokenSource := oauthCfg.TokenSource(ctx, tok)
	tokenSource := &persistingTokenSource{
		base:    baseTokenSource,
		persist: persistAuthState,
	}

	tok, err = tokenSource.Token()
	if err != nil {
		logger.Warn("oauth token refresh failed; reauthorizing", "has_refresh_token", tok.RefreshToken != "")
		tok, err = authorizeDesktopClient(ctx, oauthCfg, tok.RefreshToken == "", autoOpen)
		if err != nil {
			return nil, err
		}
		if persistAuthState {
			if err := writeCachedToken(tok); err != nil {
				return nil, err
			}
		} else {
			logger.Warn("dry run, skip oauth token cache save after reauthorization")
		}
		baseTokenSource = oauthCfg.TokenSource(ctx, tok)
		tokenSource = &persistingTokenSource{
			base:    baseTokenSource,
			persist: persistAuthState,
		}
	}

	logger.Debug("oauth http client ready")
	client := oauth2.NewClient(ctx, tokenSource)
	client.Transport = newResilientTransport(client.Transport)
	return client, nil
}

// Authorizes a desktop client using OAuth2 and returns the OAuth token
func authorizeDesktopClient(ctx context.Context, oauthCfg *oauth2.Config, forceConsent bool, autoOpen bool) (*oauth2.Token, error) {
	logger := corelog.For("firebase")
	logger.Info("start oauth desktop authorization", "force_consent", forceConsent, "auto_open", autoOpen)

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
		shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
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
	logger.Info("oauth authorization url ready", "url", authURL)

	fmt.Fprintln(os.Stderr, "Open this URL in your browser to authorize fbrcm:")
	fmt.Fprintln(os.Stderr, authURL)
	if autoOpen {
		if err := browser.OpenURL(authURL); err != nil {
			logger.Warn("open browser automatically failed", "err", err)
			fmt.Fprintf(os.Stderr, "Could not open browser automatically: %v\n", err)
		}
	} else {
		logger.Info("browser auto-open disabled")
	}
	fmt.Fprintln(os.Stderr, "Waiting for OAuth callback on local loopback server...")
	logger.Info("waiting for oauth callback")

	select {
	case code := <-codeCh:
		logger.Info("oauth callback received; exchanging code")
		tok, err := oauthCfg.Exchange(
			ctx,
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
	case <-time.After(2 * time.Minute):
		logger.Error("oauth callback timed out")
		return nil, fmt.Errorf("timed out waiting for OAuth callback")
	}
}

// Starts a local loopback server to handle OAuth2 callback requests
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
		logger.Debug("http request", "method", r.Method, "url", r.URL.String(), "headers", formatHeaders(r.Header))

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
		logger.Debug("http response headers", "method", r.Method, "url", r.URL.String(), "status", http.StatusText(http.StatusOK), "headers", formatHeaders(w.Header()))
		logger.Info("http response", "method", r.Method, "url", r.URL.String(), "status", "200 OK")
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
