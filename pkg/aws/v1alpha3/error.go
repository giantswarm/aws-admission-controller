package v1alpha3

import (
	"github.com/giantswarm/microerror"
)

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

var organizationLabelNotFoundError = &microerror.Error{
	Kind: "organizationLabelNotFoundError",
}

// IsOrganizationLabelNotFoundError asserts organizationLabelNotFoundError.
func IsOrganizationLabelNotFoundError(err error) bool {
	return microerror.Cause(err) == organizationLabelNotFoundError
}

var organizationNotFoundError = &microerror.Error{
	Kind: "organizationNotFoundError",
}

// IsOrganizationNotFoundError asserts organizationNotFoundError.
func IsOrganizationNotFoundError(err error) bool {
	return microerror.Cause(err) == organizationNotFoundError
}
