package browser

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	corelog "fbrcm/core/log"
)

// OpenURL opens url in the system browser and logs only a redacted copy of the URL.
func OpenURL(url string) error {
	logger := corelog.For("browser")
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	logger.Info("launch browser", "command", cmd.Path, "url", redactedURL(url))
	if err := cmd.Start(); err != nil {
		logger.Error("launch browser failed", "url", redactedURL(url), "err", err)
		return fmt.Errorf("launching browser: %w", err)
	}
	return nil
}

// redactedURL returns a URL string safe for logs by masking sensitive query values.
func redactedURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	query := parsed.Query()
	for key := range query {
		switch strings.ToLower(key) {
		case "access_token", "authuser", "client_secret", "code", "code_challenge", "code_verifier", "id_token", "password", "refresh_token", "state", "token":
			query.Set(key, "[REDACTED]")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
