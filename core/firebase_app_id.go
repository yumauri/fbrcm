package core

import "regexp"

var firebaseAppIDPattern = regexp.MustCompile(`([0-9]+):([0-9]+):(android|ios|web):([A-Za-z0-9]+)`)

// FirebaseAppID is the structured form of a complete Firebase application ID.
type FirebaseAppID struct {
	Version       string
	ProjectNumber string
	Platform      AppPlatform
	Hash          string
}

// ParseFirebaseAppID parses value only when the entire value is a complete
// Firebase application ID.
func ParseFirebaseAppID(value string) (FirebaseAppID, bool) {
	match := firebaseAppIDPattern.FindStringSubmatchIndex(value)
	if len(match) < 10 || match[0] != 0 || match[1] != len(value) {
		return FirebaseAppID{}, false
	}
	return FirebaseAppID{
		Version:       value[match[2]:match[3]],
		ProjectNumber: value[match[4]:match[5]],
		Platform:      AppPlatform(value[match[6]:match[7]]),
		Hash:          value[match[8]:match[9]],
	}, true
}

// FirebaseAppIDPlatformSpan identifies the platform segment of a Firebase App
// ID embedded in a larger string. Start and End are byte offsets into the
// original string.
type FirebaseAppIDPlatformSpan struct {
	Start    int
	End      int
	Platform AppPlatform
}

// FirebaseAppIDPlatformSpans finds every complete Firebase App ID in value and
// returns the platform segment of each match in source order.
func FirebaseAppIDPlatformSpans(value string) []FirebaseAppIDPlatformSpan {
	matches := firebaseAppIDPattern.FindAllStringSubmatchIndex(value, -1)
	spans := make([]FirebaseAppIDPlatformSpan, 0, len(matches))
	for _, match := range matches {
		if len(match) < 8 || !firebaseAppIDBoundary(value, match[0], match[1]) {
			continue
		}
		spans = append(spans, FirebaseAppIDPlatformSpan{
			Start:    match[6],
			End:      match[7],
			Platform: AppPlatform(value[match[6]:match[7]]),
		})
	}
	return spans
}

func firebaseAppIDBoundary(value string, start, end int) bool {
	return (start == 0 || !firebaseAppIDWordByte(value[start-1])) &&
		(end == len(value) || !firebaseAppIDWordByte(value[end]))
}

func firebaseAppIDWordByte(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' ||
		value == '_'
}
