package networkpool

import (
	"github.com/giantswarm/microerror"
)

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var intersectFailedError = &microerror.Error{
	Kind: "intersectFailedError",
}

// IsIntersectFailed asserts intersectFailedError.
func IsIntersectFailed(err error) bool {
	return microerror.Cause(err) == intersectFailedError
}

var notAllowedError = &microerror.Error{
	Kind: "notAllowedError",
}

// IsNotAllowed asserts notAllowedError.
func IsNotAllowed(err error) bool {
	return microerror.Cause(err) == notAllowedError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var parsingFailedError = &microerror.Error{
	Kind: "parsingFailedError",
}

// IsParsingFailed asserts parsingFailedError.
func IsParsingFailed(err error) bool {
	return microerror.Cause(err) == parsingFailedError
}
