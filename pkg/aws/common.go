package aws

import (
	"github.com/blang/semver"
)

const (
	// DefaultMasterReplicas is the default number of master node replicas
	DefaultMasterReplicas = 3

	// FirstHARelease is the first GS release for AWS that supports HA Masters
	FirstHARelease = "11.4.0"
)

// IsHAVersion returns whether a given releaseVersion supports HA Masters
func IsHAVersion(releaseVersion *semver.Version) bool {
	HAVersion, _ := semver.New(FirstHARelease)
	return releaseVersion.GE(*HAVersion)
}
