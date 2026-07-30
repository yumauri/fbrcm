package display

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

const localDateTimeLayout = "2006-01-02 15:04:05"

type ValueSummaryKind string

const (
	ValueSummaryPlain           ValueSummaryKind = ""
	ValueSummaryPersonalization ValueSummaryKind = "personalization"
	ValueSummaryExperiment      ValueSummaryKind = "experiment"
	ValueSummaryRollout         ValueSummaryKind = "rollout"
	ValueSummaryUnknown         ValueSummaryKind = "unknown"
)

// ValueSummary retains the structure needed to style managed Remote Config
// values consistently while keeping a plain-text representation for JSON and
// no-color output.
type ValueSummary struct {
	Kind    ValueSummaryKind
	Text    string
	Rollout *RolloutSummary
}

type RolloutSummary struct {
	Percentage string
	Value      string
}

// FormatCount formats a count with the grammatically appropriate noun.
func FormatCount(count int, singular, plural string) string {
	noun := plural
	if count == 1 {
		noun = singular
	}
	return fmt.Sprintf("%d %s", count, noun)
}

// FormatLocalDateTime formats a timestamp in the local timezone for terminal displays.
func FormatLocalDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format(localDateTimeLayout)
}

// FormatRelativeTime formats an RFC3339 timestamp relative to now.
func FormatRelativeTime(value string, now time.Time) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "1970-01-01T00:00:00") {
		return "—"
	}
	timestamp, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}

	elapsed := now.Sub(timestamp)
	future := elapsed < 0
	if future {
		elapsed = -elapsed
	}

	var relative string
	switch {
	case elapsed < time.Minute:
		relative = "less than a minute"
	case elapsed < time.Hour:
		relative = FormatCount(int(elapsed/time.Minute), "minute", "minutes")
	case elapsed < 24*time.Hour:
		relative = FormatCount(int(elapsed/time.Hour), "hour", "hours")
	case elapsed < 30*24*time.Hour:
		relative = FormatCount(int(elapsed/(24*time.Hour)), "day", "days")
	case elapsed < 365*24*time.Hour:
		relative = FormatCount(int(elapsed/(30*24*time.Hour)), "month", "months")
	default:
		relative = FormatCount(int(elapsed/(365*24*time.Hour)), "year", "years")
	}
	if future {
		return "in " + relative
	}
	return relative + " ago"
}

// FormatSummary formats a Remote Config value for tree summaries and table output.
func FormatSummary(value firebase.RemoteConfigValue, valueType string) string {
	return SummarizeValue(value, valueType).Text
}

// SummarizeValue describes a Remote Config value for human-readable renderers.
func SummarizeValue(value firebase.RemoteConfigValue, valueType string) ValueSummary {
	switch {
	case value.UseInAppDefault:
		return ValueSummary{Text: "(in-app default)"}
	case len(value.PersonalizationValue) > 0:
		return ValueSummary{Kind: ValueSummaryPersonalization, Text: "◈ (personalization)"}
	case len(value.ExperimentValue) > 0:
		return ValueSummary{Kind: ValueSummaryExperiment, Text: "⚗ (a/b test)"}
	case len(value.RolloutValue) > 0:
		return summarizeRollout(value.RolloutValue)
	case value.UnknownValueOption != "":
		return ValueSummary{Kind: ValueSummaryUnknown, Text: "(" + value.UnknownValueOption + ")"}
	case value.Value == "":
		return ValueSummary{Text: "(empty " + EmptyValueType(valueType) + ")"}
	default:
		return ValueSummary{Text: strings.ReplaceAll(value.Value, "\n", "\\n")}
	}
}

func summarizeRollout(raw json.RawMessage) ValueSummary {
	var value struct {
		Value   *string  `json:"value"`
		Percent *float64 `json:"percent"`
	}
	if err := json.Unmarshal(raw, &value); err != nil || value.Value == nil || value.Percent == nil {
		return ValueSummary{Kind: ValueSummaryRollout, Text: "(rollout)"}
	}
	displayValue := *value.Value
	if displayValue == "" {
		displayValue = `""`
	} else {
		displayValue = strings.ReplaceAll(displayValue, "\n", "\\n")
	}
	percentage := strconv.FormatFloat(*value.Percent, 'f', -1, 64) + "%"
	rollout := &RolloutSummary{Percentage: percentage, Value: displayValue}
	return ValueSummary{
		Kind:    ValueSummaryRollout,
		Text:    "◐ " + percentage + " → " + displayValue + " / ◑ (no change)",
		Rollout: rollout,
	}
}

// EmptyValueType normalizes a parameter value type for empty-value labels.
func EmptyValueType(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return "string"
	}
	return valueType
}

// FormatRawValue formats a stored raw value string for detail panels.
func FormatRawValue(value, valueType string) string {
	if value == "" {
		return "(empty " + EmptyValueType(valueType) + ")"
	}
	return strings.ReplaceAll(value, "\n", "\\n")
}
