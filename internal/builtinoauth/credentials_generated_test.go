//go:build fbrcm_google_auth

package builtinoauth

import "testing"

func TestCredentialedBuildReconstructsCompleteCredentials(t *testing.T) {
	clientID, clientSecret := Credentials()
	if clientID == "" || clientSecret == "" {
		t.Fatal("credentialed build did not reconstruct both built-in OAuth values")
	}
}
