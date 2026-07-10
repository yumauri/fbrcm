package parameters

import "regexp"

var deprecatedDescriptionPattern = regexp.MustCompile(`(?i)deprecat|obsolete|sunset|retired?|no longer|use .+ instead|replaced?|superseded?|removed?`)

func isDeprecatedDescription(description string) bool {
	return deprecatedDescriptionPattern.MatchString(description)
}
