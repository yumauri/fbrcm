package versions

import (
	"github.com/yumauri/fbrcm/core"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

type conditionBlameResult struct {
	Project             core.Project                                     `json:"project"`
	Condition           string                                           `json:"condition"`
	AtVersion           string                                           `json:"at_version"`
	ScannedVersionCount int                                              `json:"scanned_version_count"`
	HistoryExhausted    bool                                             `json:"history_exhausted"`
	Boundary            *conditionBlameBoundary                          `json:"boundary"`
	Changes             []versionBlameChangeJSON[rcdiff.ConditionChange] `json:"changes"`
}

type conditionBlameBoundary struct {
	Version string `json:"version"`
	State   string `json:"state" contract:"enum=present|absent"`
}

func conditionBlameJSON(project core.Project, history core.ConditionHistory) conditionBlameResult {
	result := conditionBlameResult{
		Project: project, Condition: history.Condition, AtVersion: history.AtVersion,
		ScannedVersionCount: history.ScannedVersionCount, HistoryExhausted: history.HistoryExhausted,
		Changes: make([]versionBlameChangeJSON[rcdiff.ConditionChange], 0, len(history.Changes)),
	}
	if history.Boundary != nil {
		state := "absent"
		if history.Boundary.Present {
			state = "present"
		}
		result.Boundary = &conditionBlameBoundary{Version: history.Boundary.Version, State: state}
	}
	for _, item := range history.Changes {
		result.Changes = append(result.Changes, blameChangeJSON(item.PreviousVersion, item.Version, item.Change))
	}
	return result
}

func renderConditionBlame(project core.Project, history core.ConditionHistory, width int) (string, error) {
	log := blameLog{
		Kind: "Condition", Name: history.Condition, Label: "Condition: " + history.Condition,
		ScannedVersionCount: history.ScannedVersionCount, HistoryExhausted: history.HistoryExhausted,
	}
	if history.Boundary != nil {
		log.Boundary = &blameLogBoundary{Version: history.Boundary.Version, Present: history.Boundary.Present}
	}

	for _, item := range history.Changes {
		log.Changes = append(log.Changes, blameLogChange{
			PreviousVersion: item.PreviousVersion, Version: item.Version,
			Diff: rcdiff.Result{Conditions: []rcdiff.ConditionChange{item.Change}},
		})
	}
	return renderBlameLog(project, log, width)
}
