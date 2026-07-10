package rc

// RemoteConfigOrder preserves the member order of an input Remote Config JSON file.
type RemoteConfigOrder struct {
	TopLevel          []string
	Parameters        []string
	Groups            []string
	GroupParameters   map[string][]string
	ConditionalValues map[string][]string
	VersionRaw        []byte
}

func orderPath(groupName, paramKey string) string {
	if groupName == "" {
		return paramKey
	}
	return groupName + "\x00" + paramKey
}
