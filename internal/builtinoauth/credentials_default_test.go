//go:build !fbrcm_google_auth && !fbrcm_e2e

package builtinoauth

import "testing"

func TestOrdinaryBuildHasNoBuiltInCredentials(t *testing.T) {
	clientID, clientSecret := Credentials()
	if clientID != "" || clientSecret != "" {
		t.Fatal("ordinary source build unexpectedly contains built-in OAuth credentials")
	}
}
