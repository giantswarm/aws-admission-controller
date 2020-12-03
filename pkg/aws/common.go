package aws

import (
	"strconv"

	"github.com/blang/semver"
	"github.com/dylanmei/iso8601"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// DefaultClusterDescription is the default name for a cluster
	DefaultClusterDescription = "Unnamed cluster"

	// DefaultMasterReplicas is the default number of master node replicas
	DefaultMasterReplicas = 3

	// DefaultNodePoolAZs is the default number of AZs used in a nodepool
	DefaultNodePoolAZs = 1

	// DefaultMasterInstanceType is the default master instance type
	DefaultMasterInstanceType = "m5.xlarge"

	// FirstHARelease is the first GS release for AWS that supports HA Masters
	FirstHARelease = "11.4.0"
)

const (
	// annotations should  taken from https://github.com/giantswarm/apiextensions/blob/master/pkg/annotation/aws.go
	// once the service is migrate to apiextensions v3
	AnnotationUpdateMaxBatchSize = "alpha.aws.giantswarm.io/update-max-batch-size"
	AnnotationUpdatePauseTime    = "alpha.aws.giantswarm.io/update-pause-time"
)

// DefaultCredentialSecret returns the default credentials for clusters
func DefaultCredentialSecret() types.NamespacedName {
	return types.NamespacedName{
		Name:      "credential-default",
		Namespace: "giantswarm",
	}
}

// ValidMasterReplicas are the allowed number of master node replicas
func ValidMasterReplicas() []int {
	return []int{1, 3}
}

// IsHAVersion returns whether a given releaseVersion supports HA Masters
func IsHAVersion(releaseVersion *semver.Version) bool {
	HAVersion, _ := semver.New(FirstHARelease)
	return releaseVersion.GE(*HAVersion)
}

// IsVersionProductionReady returns whether a given releaseVersion is not a prerelease or test version
func IsVersionProductionReady(version *semver.Version) bool {
	return len(version.Pre) == 0 && len(version.Build) == 0
}

// IsValidMasterReplicas returns whether a given number is a valid number of Master node replicas
func IsValidMasterReplicas(replicas int) bool {
	for _, r := range ValidMasterReplicas() {
		if r == replicas {
			return true
		}
	}
	return false
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
	ratio, err := strconv.ParseFloat(value, 10)
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
