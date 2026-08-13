package config

// InvalidConfigurationError identifies malformed persisted configuration
// separately from filesystem failures and missing optional files.
type InvalidConfigurationError struct {
	Path  string
	Stage string
	Err   error
}

func (e *InvalidConfigurationError) Error() string { return e.Err.Error() }
func (e *InvalidConfigurationError) Unwrap() error { return e.Err }

func invalidConfiguration(path, stage string, err error) error {
	return &InvalidConfigurationError{Path: path, Stage: stage, Err: err}
}
