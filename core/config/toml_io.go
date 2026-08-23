package config

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

func decodeTOMLWithOptions(data []byte, dest any, strict bool) error {
	decoder := toml.NewDecoder(bytes.NewReader(data))
	if strict {
		decoder.DisallowUnknownFields()
	}
	err := decoder.Decode(dest)
	if strictErr, ok := errors.AsType[*toml.StrictMissingError](err); ok {
		return fmt.Errorf("%w:\n%s", err, strictErr.String())
	}
	return err
}

func encodeTOML(v any) ([]byte, error) {
	return toml.Marshal(v)
}
