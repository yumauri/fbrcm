package diffinput

import (
	"encoding/json"
	"strconv"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
)

// Condition prepares one Remote Config condition as a generic dictionary.
func Condition(position int, condition *firebase.RemoteConfigCondition) dictdiff.Dictionary {
	if condition == nil {
		return dictdiff.Dictionary{}
	}
	return dictdiff.Dictionary{
		"position":   dictdiff.Number(json.Number(strconv.Itoa(position))),
		"expression": dictdiff.String(condition.Expression),
		"color":      dictdiff.Enum(condition.TagColor),
	}
}

// Group prepares one Remote Config parameter group as a generic dictionary.
// present distinguishes an absent group from one with an empty description.
func Group(description string, present bool) dictdiff.Dictionary {
	if !present {
		return dictdiff.Dictionary{}
	}
	return dictdiff.Dictionary{
		"description": dictdiff.String(description),
	}
}
