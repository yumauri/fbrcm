package harness

import "fmt"

type Mode string

const (
	ModeReplay        Mode = "replay"
	ModeRecordMissing Mode = "record-missing"
	ModeRefreshHTTP   Mode = "refresh-http"
	ModeUpdateOutput  Mode = "update-output"
	ModeRefreshAll    Mode = "refresh-all"
)

func ParseMode(value string) (Mode, error) {
	mode := Mode(value)
	switch mode {
	case ModeReplay, ModeRecordMissing, ModeRefreshHTTP, ModeUpdateOutput, ModeRefreshAll:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported E2E mode %q", value)
	}
}

func (m Mode) Capture(cassetteExists bool) bool {
	switch m {
	case ModeRecordMissing:
		return !cassetteExists
	case ModeRefreshHTTP, ModeRefreshAll:
		return true
	default:
		return false
	}
}

func (m Mode) UpdateOutput(snapshotExists bool) bool {
	if m == ModeRefreshAll || m == ModeUpdateOutput {
		return true
	}
	return m == ModeRecordMissing && !snapshotExists
}
