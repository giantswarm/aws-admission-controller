package aws

import (
	"github.com/blang/semver"
)

const (
	// DefaultMasterReplicas is the default number of master node replicas
	DefaultMasterReplicas = 3

	// DefaultMasterInstanceType is the default master instance type
	DefaultMasterInstanceType = "m5.xlarge"

	// FirstHARelease is the first GS release for AWS that supports HA Masters
	FirstHARelease = "11.4.0"

	// CreateOperation is the string attribute in an admission request for creation
	CreateOperation = "CREATE"
)

// ValidMasterReplicas are the allowed number of master node replicas
func ValidMasterReplicas() []int {
	return []int{1, 3}
}

// IsHAVersion returns whether a given releaseVersion supports HA Masters
func IsHAVersion(releaseVersion *semver.Version) bool {
	HAVersion, _ := semver.New(FirstHARelease)
	return releaseVersion.GE(*HAVersion)
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
