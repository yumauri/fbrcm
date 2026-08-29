//go:build !fbrcm_google_auth && !fbrcm_e2e

package builtinoauth

// Credentials returns no built-in client for ordinary source builds.
func Credentials() (clientID, clientSecret string) {
	return "", ""
}
