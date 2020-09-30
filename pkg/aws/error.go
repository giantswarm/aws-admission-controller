package aws

import (
	"github.com/giantswarm/microerror"
)

var ExecutionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == ExecutionFailedError
}

var NotAllowedError = &microerror.Error{
	Kind: "notAllowedError",
}

// IsNotAllowed asserts notAllowedError.
func IsNotAllowed(err error) bool {
	return microerror.Cause(err) == NotAllowedError
}

var NotFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == NotFoundError
}

var InvalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == InvalidConfigError
}

var ParsingFailedError = &microerror.Error{
	Kind: "parsingFailedError",
}

// IsParsingFailed asserts parsingFailedError.
func IsParsingFailed(err error) bool {
	return microerror.Cause(err) == ParsingFailedError
}

var ControlPlaneLabelNotEqualError = &microerror.Error{
	Kind: "controlPlaneLabelNotEqualError",
}

// IsControlPlaneLabelNotEqualError asserts controlPlaneLabelNotEqualError.
func IsControlPlaneLabelNotEqualError(err error) bool {
	return microerror.Cause(err) == ControlPlaneLabelNotEqualError
}
