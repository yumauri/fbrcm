package firebase

// RemoteConfigConsoleURL handles remote config console url and returns the resulting value or error.
func RemoteConfigConsoleURL(projectID string) string {
	return "https://console.firebase.google.com/project/" + projectID + "/config"
}
