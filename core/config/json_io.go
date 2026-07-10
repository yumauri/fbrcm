package config

import (
	"encoding/json"
	"errors"
	"os"
)

type decodeError struct {
	err error
}

func (e *decodeError) Error() string {
	return e.err.Error()
}

func (e *decodeError) Unwrap() error {
	return e.err
}

type encodeError struct {
	err error
}

func (e *encodeError) Error() string {
	return e.err.Error()
}

func (e *encodeError) Unwrap() error {
	return e.err
}

func isNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

func isDecodeError(err error) bool {
	var decodeErr *decodeError
	return errors.As(err, &decodeErr)
}

func isEncodeError(err error) bool {
	var encodeErr *encodeError
	return errors.As(err, &encodeErr)
}

func readFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func decodeJSON(data []byte, dest any) error {
	return json.Unmarshal(data, dest)
}

func readJSONFile(path string, dest any) error {
	data, err := readFileBytes(path)
	if err != nil {
		return err
	}
	if err := decodeJSON(data, dest); err != nil {
		return &decodeError{err: err}
	}
	return nil
}

func encodeJSONIndent(v any) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func writeJSONFile(path string, v any) error {
	data, err := encodeJSONIndent(v)
	if err != nil {
		return &encodeError{err: err}
	}
	return WritePrivateFile(path, data)
}
