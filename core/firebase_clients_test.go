package core

import (
	"errors"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestWithAuthFailureIDClassifiesQuotaProjectSources(t *testing.T) {
	for _, test := range []struct {
		name     string
		source   firebase.QuotaProjectSource
		wantKind string
	}{
		{name: "environment", source: firebase.QuotaProjectSourceEnvironment, wantKind: "configuration"},
		{name: "credentials", source: firebase.QuotaProjectSourceCredentials, wantKind: firebase.AuthenticationCredentialsInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			cause := &firebase.QuotaProjectError{Source: test.source, Err: errors.New("invalid quota project")}
			err := withAuthFailureID("personal", cause)
			var authErr *AuthError
			if !errors.As(err, &authErr) || authErr.Kind != test.wantKind || authErr.AuthID != "personal" || !errors.Is(err, cause) {
				t.Fatalf("error = %#v, want auth kind %q", err, test.wantKind)
			}
		})
	}
}
