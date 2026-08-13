package firebase

import "testing"

func TestCredentialValidationMatchesPublishedRequiredFields(t *testing.T) {
	oauth := []byte(`{
		"installed": {
			"client_id": "client",
			"client_secret": "secret",
			"auth_uri": "https://accounts.example/authorize",
			"token_uri": "https://accounts.example/token",
			"redirect_uris": ["http://localhost"]
		}
	}`)
	if err := ValidateOAuthClientSecret(oauth); err != nil {
		t.Fatalf("valid OAuth credentials = %v", err)
	}
	if err := ValidateOAuthClientSecret([]byte(`{"installed":{"redirect_uris":["http://localhost"]}}`)); err == nil {
		t.Fatal("OAuth credentials without required fields were accepted")
	}

	serviceAccount := []byte(`{
		"type": "service_account",
		"project_id": "demo",
		"private_key": "key",
		"client_email": "service@example.com",
		"token_uri": "https://accounts.example/token"
	}`)
	if err := ValidateServiceAccountKey(serviceAccount); err != nil {
		t.Fatalf("valid service-account credentials = %v", err)
	}
	if err := ValidateServiceAccountKey([]byte(`{"type":"service_account"}`)); err == nil {
		t.Fatal("service-account credentials without required fields were accepted")
	}
}
