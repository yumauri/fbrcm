package shared

import "fmt"

// ExitError lets commands request a specific process exit status.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error { return e.Err }

func WithExitCode(err error, code int) error {
	if err == nil {
		return &ExitError{Code: code}
	}
	return &ExitError{Code: code, Err: fmt.Errorf("%w", err)}
}
