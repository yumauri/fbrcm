package firebase

import (
	"context"

	"fbrcm/core/config"
)

type WhoAmI struct {
	SecretPath  string `json:"secret_path"`
	TokenPath   string `json:"token_path"`
	TokenExpiry string `json:"token_expiry"`
}

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
