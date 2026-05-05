package firebase

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"

	"fbrcm/core/config"
	corelog "fbrcm/core/log"
)

// oauth2.TokenSource implementation with persistent caching of OAuth tokens
type persistingTokenSource struct {
	base      oauth2.TokenSource
	lastToken *oauth2.Token
	persist   bool
}

// Returns an OAuth token, caching it to disk if it has changed
func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	logger := corelog.For("firebase")
	tok, err := p.base.Token()
	if err != nil {
		logger.Error("oauth token source failed", "err", err)
		return nil, err
	}

	if !tokensEqual(p.lastToken, tok) {
		if p.persist {
			logger.Info("oauth token changed; persist cache")
			if err := writeCachedToken(tok); err != nil {
				return nil, err
			}
		} else {
			logger.Warn("dry run, skip oauth token cache update")
		}
		p.lastToken = tok
	}

	return tok, nil
}

// Compares two OAuth tokens for equality
func tokensEqual(a, b *oauth2.Token) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.AccessToken == b.AccessToken &&
		a.RefreshToken == b.RefreshToken &&
		a.TokenType == b.TokenType &&
		a.Expiry.Equal(b.Expiry)
}

// Reads the cached OAuth token from disk
func readCachedToken() (*oauth2.Token, error) {
	path := config.GetTokenFilePath()
	logger := corelog.For("firebase")
	logger.Debug("read cached oauth token", "path", path)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("oauth token cache miss", "path", path)
			return nil, nil
		}
		logger.Error("read cached oauth token failed", "path", path, "err", err)
		return nil, fmt.Errorf("reading cached OAuth token: %w", err)
	}

	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		logger.Error("decode cached oauth token failed", "path", path, "err", err)
		return nil, fmt.Errorf("decoding cached OAuth token: %w", err)
	}
	logger.Info("loaded cached oauth token", "path", path, "has_refresh_token", tok.RefreshToken != "")
	return &tok, nil
}

// Writes the OAuth token to disk for caching
func writeCachedToken(tok *oauth2.Token) error {
	path := config.GetTokenFilePath()
	logger := corelog.For("firebase")
	if err := config.EnsurePrivateDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("creating token cache directory: %w", err)
	}

	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		logger.Error("encode oauth token failed", "path", path, "err", err)
		return fmt.Errorf("encoding OAuth token: %w", err)
	}
	logger.Debug("write cached oauth token", "path", path)
	if err := os.WriteFile(path, data, config.PrivateFileMode); err != nil {
		logger.Error("write cached oauth token failed", "path", path, "err", err)
		return fmt.Errorf("writing cached OAuth token: %w", err)
	}
	if err := config.EnsurePrivateFile(path); err != nil {
		return fmt.Errorf("tightening cached OAuth token permissions: %w", err)
	}
	logger.Debug("saved cached oauth token", "path", path)
	return nil
}

func ReadCachedToken() (*oauth2.Token, error) {
	return readCachedToken()
}
