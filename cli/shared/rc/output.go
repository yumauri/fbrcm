package rc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

// WriteRemoteConfigFile writes normalized Remote Config JSON to a private file.
func WriteRemoteConfigFile(path string, raw []byte) error {
	raw = TrimTrailingLineBreaks(NormalizeExportJSON(raw))
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create destination dir: %w", err)
		}
	}
	if err := config.WritePrivateFile(path, raw); err != nil {
		return fmt.Errorf("write destination file: %w", err)
	}
	return nil
}

// OrderMutator adjusts member order after a stdin mutation.
type OrderMutator func(order *RemoteConfigOrder)

// WriteOrderPreservingRemoteConfigStdout writes finalCfg to stdout using member order from raw input.
func WriteOrderPreservingRemoteConfigStdout(cmd *cobra.Command, finalCfg *firebase.RemoteConfig, remoteConfigRaw []byte) error {
	return WriteOrderPreservingRemoteConfigStdoutWithOrder(cmd, finalCfg, remoteConfigRaw, nil)
}

// WriteOrderPreservingRemoteConfigStdoutWithOrder writes finalCfg to stdout using member order
// from raw input, optionally adjusted by mutate.
func WriteOrderPreservingRemoteConfigStdoutWithOrder(cmd *cobra.Command, finalCfg *firebase.RemoteConfig, remoteConfigRaw []byte, mutate OrderMutator) error {
	order, err := ParseRemoteConfigOrder(remoteConfigRaw)
	if err != nil {
		return fmt.Errorf("parse stdin remote config order: %w", err)
	}
	if mutate != nil {
		mutate(&order)
	}
	out, err := MarshalPrettyRemoteConfigWithOrder(finalCfg, order)
	if err != nil {
		return err
	}
	if _, err := cmd.OutOrStdout().Write(out); err != nil {
		return err
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}
