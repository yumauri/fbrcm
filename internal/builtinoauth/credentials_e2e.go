//go:build fbrcm_e2e && !fbrcm_google_auth

package builtinoauth

// Credentials supplies deliberately invalid, non-sensitive values for the E2E
// suite. Production credentialed builds use the generated implementation.
func Credentials() (clientID, clientSecret string) {
	return "e2e-client-id", "e2e-client-secret"
}
