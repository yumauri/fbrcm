package shared

import "github.com/yumauri/fbrcm/cli/machine"

type Remediation = machine.Remediation
type ArgumentError = machine.ArgumentError
type ExpressionError = machine.ExpressionError
type ValidationError = machine.ValidationError
type ConflictError = machine.ConflictError
type SelectionError = machine.SelectionError
type SelectionCandidate = machine.SelectionCandidate
type ProjectResolutionError = machine.SelectionError
type BatchError = machine.BatchError
type BatchFailure = machine.BatchFailure

const (
	RemediationRetryWithArguments = machine.RemediationRetryWithArguments
	RemediationReplaceSelector    = machine.RemediationReplaceSelector
	RemediationRunCommand         = machine.RemediationRunCommand
)

func InvalidArgument(err error) error { return machine.InvalidArgument(err) }
func InvalidInput(code, source string, err error) error {
	return machine.InvalidInput(code, source, err)
}

func SafeText(value string) string   { return machine.SafeText(value) }
func SafeErrorText(err error) string { return machine.SafeErrorText(err) }
