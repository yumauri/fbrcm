// Package machine contains CLI-contract primitives that must be shared by
// command, orchestration, and rendering packages without creating import cycles.
package machine

import (
	"fmt"
)

type Remediation struct {
	Description string
	Strategy    string
	Argv        []string
}

const (
	RemediationRetryWithArguments = "retry_with_arguments"
	RemediationReplaceSelector    = "replace_selector"
	RemediationRunCommand         = "run_command"
)

type ArgumentError struct {
	Code string
	Err  error
}

func (e *ArgumentError) Error() string { return e.Err.Error() }
func (e *ArgumentError) Unwrap() error { return e.Err }

func InvalidArgument(err error) error {
	if err == nil {
		return nil
	}
	return &ArgumentError{Code: "argument.invalid", Err: err}
}

type ExpressionError struct {
	Expression string
	Context    string
	Target     string
	Err        error
}

func (e *ExpressionError) Error() string { return e.Err.Error() }
func (e *ExpressionError) Unwrap() error { return e.Err }

type ValidationError struct {
	Code   string
	Source string
	Stage  string
	Target string
	Err    error
}

func (e *ValidationError) Error() string { return e.Err.Error() }
func (e *ValidationError) Unwrap() error { return e.Err }

func InvalidInput(code, source string, err error) error {
	if err == nil {
		return nil
	}
	return &ValidationError{Code: code, Source: source, Stage: "input", Err: err}
}

type ConflictError struct {
	Code        string
	Resource    string
	Target      string
	Retryable   bool
	Remediation []Remediation
	Err         error
}

func (e *ConflictError) Error() string { return e.Err.Error() }
func (e *ConflictError) Unwrap() error { return e.Err }

type SelectionCandidate struct {
	Name string
	ID   string
}

type SelectionError struct {
	Resource   string
	Kind       string
	Query      string
	Candidates []SelectionCandidate
	Err        error
}

func (e *SelectionError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	switch e.Kind {
	case "ambiguous":
		return fmt.Sprintf("several %ss match %q", e.Resource, e.Query)
	default:
		return fmt.Sprintf("%s %q was not found", e.Resource, e.Query)
	}
}

func (e *SelectionError) Unwrap() error { return e.Err }

type BatchError struct {
	Operation             string
	FailedTargets         []string
	Failures              []BatchFailure
	SuccessfulTargetCount int
	PublishedTargetCount  int
	Remediation           []Remediation
	Err                   error
}

// BatchFailure preserves the typed cause for one failed target so machine
// output can expose actionable per-target problem semantics.
type BatchFailure struct {
	Target string
	Err    error
}

func (e *BatchError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("%d batch targets failed", len(e.FailedTargets))
}

func (e *BatchError) Unwrap() error { return e.Err }
