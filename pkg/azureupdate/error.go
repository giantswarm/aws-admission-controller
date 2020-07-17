package azureupdate

import (
	"github.com/giantswarm/microerror"
)

var invalidOperationError = &microerror.Error{
	Kind: "invalidOperationError",
}

// IsInvalidOperationError asserts invalidOperationError.
func IsInvalidOperationError(err error) bool {
	return microerror.Cause(err) == invalidOperationError
}

var invalidReleaseError = &microerror.Error{
	Kind: "invalidReleaseError",
}

// IsInvalidReleaseError asserts parsingFailedError.
func IsInvalidReleaseError(err error) bool {
	return microerror.Cause(err) == invalidReleaseError
}

var parsingFailedError = &microerror.Error{
	Kind: "parsingFailedError",
}

// IsParsingFailed asserts parsingFailedError.
func IsParsingFailed(err error) bool {
	return microerror.Cause(err) == parsingFailedError
}

var unknownReleaseError = &microerror.Error{
	Kind: "unknownReleaseError",
}

// IsUnknownReleaseError asserts parsingFailedError.
func IsUnknownReleaseError(err error) bool {
	return microerror.Cause(err) == unknownReleaseError
}
