package firebase

import (
	"context"

	"fbrcm/core/config"
)

// WhoAmI holds who am i state used by the firebase package.
type WhoAmI struct {
	// SecretPath stores secret path for WhoAmI.
	SecretPath string `json:"secret_path"`
	// TokenPath stores token path for WhoAmI.
	TokenPath string `json:"token_path"`
	// TokenExpiry stores token expiry for WhoAmI.
	TokenExpiry string `json:"token_expiry"`
}

// ReadWhoAmI reads who am i and returns the resulting value or error.
func ReadWhoAmI(ctx context.Context) (*WhoAmI, error) {
	_ = ctx

	info := &WhoAmI{
		SecretPath: config.GetSecretFilePath(),
	}

	tok, err := ReadCachedToken()
	if err != nil {
		return nil, err
	}
	if tok == nil {
		return info, nil
	}

	info.TokenPath = config.GetTokenFilePath()
	if !tok.Expiry.IsZero() {
		info.TokenExpiry = tok.Expiry.UTC().Format("2006-01-02T15:04:05Z07:00")
	}

	return info, nil
}
