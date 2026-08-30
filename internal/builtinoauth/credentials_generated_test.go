//go:build fbrcm_google_auth

package builtinoauth

import "testing"

func TestGeneratedBuildHasCompleteCredentialPair(t *testing.T) {
	clientID, clientSecret := Credentials()
	if (clientID == "") != (clientSecret == "") {
		t.Fatal("generated build contains only one built-in OAuth value")
	}
}
