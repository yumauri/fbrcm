package rc

import (
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/machine"
	"github.com/yumauri/fbrcm/ops/shared/fileoutput"
)

// WriteRemoteConfigFile writes normalized Remote Config JSON to a private file.
func WriteRemoteConfigFile(path string, raw []byte) error {
	return writeRemoteConfigFile(path, raw, false)
}

// CreateRemoteConfigFile writes normalized Remote Config JSON without
// replacing an existing destination.
func CreateRemoteConfigFile(path string, raw []byte) error {
	return writeRemoteConfigFile(path, raw, true)
}

func writeRemoteConfigFile(path string, raw []byte, exclusive bool) error {
	raw = NormalizeExportBytes(raw)
	write := fileoutput.Write
	if exclusive {
		write = fileoutput.Create
	}
	return write(path, raw)
}

// NormalizeExportBytes returns the exact stable bytes written by Remote Config
// export commands and described by their artifact metadata.
func NormalizeExportBytes(raw []byte) []byte {
	return TrimTrailingLineBreaks(NormalizeExportJSON(raw))
}

// OrderMutator adjusts member order after a stdin mutation.
type OrderMutator func(order *RemoteConfigOrder)

// WriteOrderPreservingRemoteConfigStdout writes finalCfg to stdout using member order from raw input.
func WriteOrderPreservingRemoteConfigStdout(cmd invocation.Call, finalCfg *firebase.RemoteConfig, remoteConfigRaw []byte) error {
	return WriteOrderPreservingRemoteConfigStdoutWithOrder(cmd, finalCfg, remoteConfigRaw, nil)
}

// WriteOrderPreservingRemoteConfigStdoutWithOrder writes finalCfg to stdout using member order
// from raw input, optionally adjusted by mutate.
func WriteOrderPreservingRemoteConfigStdoutWithOrder(cmd invocation.Call, finalCfg *firebase.RemoteConfig, remoteConfigRaw []byte, mutate OrderMutator) error {
	out, err := MarshalOrderPreservingRemoteConfig(finalCfg, remoteConfigRaw, mutate)
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

// MarshalOrderPreservingRemoteConfig returns the transformed JSON bytes using
// member order from the raw input, optionally adjusted by mutate.
func MarshalOrderPreservingRemoteConfig(finalCfg *firebase.RemoteConfig, remoteConfigRaw []byte, mutate OrderMutator) ([]byte, error) {
	order, err := ParseRemoteConfigOrder(remoteConfigRaw)
	if err != nil {
		return nil, machine.InvalidInput("stdin.remote_config.invalid", "stdin", fmt.Errorf("parse stdin remote config order: %w", err))
	}
	if mutate != nil {
		mutate(&order)
	}
	out, err := MarshalPrettyRemoteConfigWithOrder(finalCfg, order)
	if err != nil {
		return nil, err
	}
	return out, nil
}
