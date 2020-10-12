package g8scontrolplane

import (
	"github.com/giantswarm/microerror"
)

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func isExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var notAllowedError = &microerror.Error{
	Kind: "notAllowedError",
}

// IsNotAllowed asserts notAllowedError.
func isNotAllowed(err error) bool {
	return microerror.Cause(err) == notAllowedError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func isNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func isInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var parsingFailedError = &microerror.Error{
	Kind: "parsingFailedError",
}

// IsParsingFailed asserts parsingFailedError.
func isParsingFailed(err error) bool {
	return microerror.Cause(err) == parsingFailedError
}

var controlPlaneLabelNotEqualError = &microerror.Error{
	Kind: "controlPlaneLabelNotEqualError",
}

// IsControlPlaneLabelNotEqualError asserts controlPlaneLabelNotEqualError.
func isControlPlaneLabelNotEqualError(err error) bool {
	return microerror.Cause(err) == controlPlaneLabelNotEqualError
}
