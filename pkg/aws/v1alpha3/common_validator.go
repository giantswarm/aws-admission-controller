package v1alpha3

import (
	"context"
	"fmt"
	"strconv"

	"github.com/blang/semver/v4"
	"github.com/dylanmei/iso8601"
	"github.com/giantswarm/microerror"
	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-admission-controller/v4/pkg/internal/normalize"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/mutator"
)

func ValidateLabelKeys(m *Handler, old metav1.Object, new metav1.Object) error {
	// validate for each giantswarm.io label that its value has not been modified
	oldLabels := old.GetLabels()
	newLabels := new.GetLabels()
	for key := range oldLabels {
		if !IsGiantSwarmLabel(key) || IsProviderTagLabel(key) || IsServicePriorityLabel(key) {
			continue
		}
		if _, ok := newLabels[key]; !ok {
			return microerror.Maskf(notAllowedError, fmt.Sprintf("User is not allowed to rename or delete label key %s.",
				key),
			)
		}
	}

	return nil
}

func ValidateLabelValues(m *Handler, old metav1.Object, new metav1.Object) error {
	// validate for each non-version label that its value has not been modified
	oldLabels := old.GetLabels()
	newLabels := new.GetLabels()
	for key, value := range oldLabels {
		if IsVersionLabel(key) || IsProviderTagLabel(key) || !IsGiantSwarmLabel(key) || IsServicePriorityLabel(key) {
			continue
		}
		if value != newLabels[key] {
			return microerror.Maskf(notAllowedError, fmt.Sprintf("User is not allowed to change label %s value from %v to %v.",
				key,
				value,
				newLabels[key]),
			)
		}
	}

	return nil
}

func ValidateOrgNamespace(meta metav1.Object) error {
	releaseVersion, err := ReleaseVersion(meta, []mutator.PatchOperation{})
	if err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse release version from object")
	}

	if !IsOrgNamespaceVersion(releaseVersion) {
		return nil
	}

	organization := key.Organization(meta)
	if organization == "" {
		return microerror.Maskf(organizationLabelNotFoundError, "Object %s Organization label %#q is empty.", meta.GetName(), label.Organization)
	}

	if !isOrgNamespace(meta.GetNamespace(), organization) {
		return microerror.Maskf(notAllowedError, "Object %s is in invalid namespace %s. Valid namespace for organization %s is %s.",
			meta.GetName(),
			meta.GetNamespace(),
			organization,
			key.OrganizationNamespaceFromName(organization))
	}

	return nil
}

func ValidateOperatorVersion(meta metav1.Object) error {
	var labels = meta.GetLabels()
	var err error

	if version, exists := labels[label.AWSOperatorVersion]; exists {
		if _, err = semver.New(version); err != nil {
			return microerror.Maskf(notAllowedError, "Object %s has invalid aws-operator version %s.",
				meta.GetName(),
				version)
		}
	}

	if version, exists := labels[label.ClusterOperatorVersion]; exists {
		if _, err = semver.New(version); err != nil {
			return microerror.Maskf(notAllowedError, "Object %s has invalid cluster-operator version %s.",
				meta.GetName(),
				version)
		}
	}

	return nil
}

func isOrgNamespace(namespace string, organization string) bool {
	return namespace == key.OrganizationNamespaceFromName(organization)
}

// MaxBatchSizeIsValid will validate the value into valid maxBatchSize
// valid values can be either:
// an integer bigger than 0
// a float between 0 < x <= 1
// float value is used as ratio of a total worker count
func MaxBatchSizeIsValid(value string) bool {
	// try parse an integer
	integer, err := strconv.Atoi(value)
	if err == nil {
		// check if the value is bigger than zero
		if integer > 0 {
			// integer value can be directly used, no need for any adjustment
			return true
		} else {
			// the value is outside of valid bounds, it cannot be used
			return false
		}
	}
	// try parse float
	ratio, err := strconv.ParseFloat(value, 32)
	if err != nil {
		// not integer or float which means invalid value
		return false
	}
	// valid value is a decimal representing a percentage
	// anything smaller than 0 or bigger than 1 is not valid
	if ratio > 0 && ratio <= 1.0 {
		return true
	}

	return false
}

// PauseTimeIsValid checks if the value is in proper ISO 8601 duration format
// and ensure the duration is not bigger than 1 Hour (AWS limitation)
func PauseTimeIsValid(value string) bool {
	d, err := iso8601.ParseDuration(value)
	if err != nil {
		return false
	}

	if d.Hours() > 1.0 {
		// AWS allows maximum of 1 hour
		return false
	}

	return true
}

func IsIntegerGreaterThanZero(v string) bool {
	// try parse an integer
	integer, err := strconv.Atoi(v)
	if err == nil {
		// check if the value is bigger than zero
		if integer > 0 {
			return true
		}
	}
	// the value is outside of valid bounds
	return false
}

func IsValidAvailabilityZones(availabilityZones []string, validAvailabilityZones []string) bool {
	if len(availabilityZones) == 0 && len(validAvailabilityZones) > 0 {
		return false
	}
	for _, az := range availabilityZones {
		if !Contains(validAvailabilityZones, az) {
			return false
		}
	}
	return true
}
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ValidateOrganizationLabelContainsExistingOrganization(ctx context.Context, ctrlClient client.Client, obj metav1.Object) error {
	organizationName, ok := obj.GetLabels()[label.Organization]
	if !ok {
		return microerror.Maskf(organizationLabelNotFoundError, "CR doesn't contain Organization label %#q", label.Organization)
	}

	organization := &securityv1alpha1.Organization{}
	err := ctrlClient.Get(ctx, client.ObjectKey{Name: normalize.AsDNSLabelName(organizationName)}, organization)
	if apierrors.IsNotFound(err) {
		return microerror.Maskf(organizationNotFoundError, "Organization label %#q must contain an existing organization, got %#q but didn't find any CR with name %#q", label.Organization, organizationName, normalize.AsDNSLabelName(organizationName))
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
