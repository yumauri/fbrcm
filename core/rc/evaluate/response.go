package evaluate

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func readCapturedResponse(exchange *Exchange, body []byte) {
	var payload struct {
		Entries         map[string]string
		State           string
		Version         json.RawMessage            `json:"templateVersion"`
		Personalization map[string]json.RawMessage `json:"personalizationMetadata"`
		Rollouts        []struct {
			ID         string   `json:"rolloutId"`
			Variant    string   `json:"variantId"`
			Parameters []string `json:"affectedParameterKeys"`
		} `json:"rolloutMetadata"`
		Experiments []struct {
			ID string `json:"experimentId"`
		} `json:"experimentDescriptions"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return
	}
	exchange.ResponseState = payload.State
	if len(payload.Version) > 0 && string(payload.Version) != "null" {
		var number json.Number
		if json.Unmarshal(payload.Version, &number) == nil {
			exchange.Version = number.String()
		}
	}
	switch payload.State {
	case "NO_CHANGE":
		exchange.Status = http.StatusNotModified
		return
	case "NO_TEMPLATE", "EMPTY_CONFIG":
		payload.Entries = map[string]string{}
	case "INSTANCE_STATE_UNSPECIFIED":
		exchange.Status = 500
		return
	}
	if payload.Entries != nil {
		exchange.Entries, exchange.HasResponse = payload.Entries, true
	}
	exchange.Managed = make(map[string]string)
	exchange.Assignments = make(map[string][]Assignment)
	for key := range payload.Personalization {
		exchange.Managed[key] = "personalization"
		exchange.Assignments[key] = append(exchange.Assignments[key], Assignment{Kind: "personalization"})
	}
	for _, rollout := range payload.Rollouts {
		for _, key := range rollout.Parameters {
			id, variant := rollout.ID, rollout.Variant
			exchange.Assignments[key] = append(exchange.Assignments[key], Assignment{Kind: "rollout", ID: &id, Variant: &variant})
			exchange.Managed[key] = fmt.Sprintf("rollout %s, variant %s", rollout.ID, rollout.Variant)
		}
	}
	for _, experiment := range payload.Experiments {
		exchange.Experiments = append(exchange.Experiments, experiment.ID)
	}
}
