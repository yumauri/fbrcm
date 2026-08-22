package firebase

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode"

	"golang.org/x/oauth2"
)

const accessTokenAuthType = "access-token"

func accessTokenHTTPClient(ctx context.Context, accessToken string) (*http.Client, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, credentialAuthenticationError(accessTokenAuthType, "load_credentials", errors.New("access token is empty"))
	}
	if strings.IndexFunc(accessToken, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return nil, credentialAuthenticationError(accessTokenAuthType, "load_credentials", errors.New("access token contains whitespace or control characters"))
	}

	token := &oauth2.Token{AccessToken: accessToken, TokenType: "Bearer"}
	return wrapAuthHTTPClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))), nil
}
