package config

import (
	"github.com/pelletier/go-toml/v2"
)

func decodeTOML(data []byte, dest any) error {
	return toml.Unmarshal(data, dest)
}

func encodeTOML(v any) ([]byte, error) {
	return toml.Marshal(v)
}

func readTOMLFile(path string, dest any) error {
	data, err := readFileBytes(path)
	if err != nil {
		return err
	}
	if err := decodeTOML(data, dest); err != nil {
		return &decodeError{err: err}
	}
	return nil
}

func writeTOMLFile(path string, v any) error {
	data, err := encodeTOML(v)
	if err != nil {
		return &encodeError{err: err}
	}
	return WritePrivateFile(path, data)
}
