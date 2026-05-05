package browser

import (
	"fmt"
	"os/exec"
	"runtime"

	corelog "fbrcm/core/log"
)

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

	logger.Info("launch browser", "command", cmd.String(), "url", url)
	if err := cmd.Start(); err != nil {
		logger.Error("launch browser failed", "url", url, "err", err)
		return fmt.Errorf("launching browser: %w", err)
	}
	return nil
}
