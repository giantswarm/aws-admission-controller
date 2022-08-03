package v1alpha3

import (
	"strings"

	"github.com/blang/semver/v4"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
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

	// FirstCAPIRelease is the first GS release that runs on CAPI controllers
	FirstCAPIRelease = "20.0.0-alpha1"

	// FirstHARelease is the first GS release for AWS that supports HA Masters
	FirstHARelease = "11.4.0"

	// FirstV1Alpha3Release is the first GS release for v1alpha3 GiantSwarm AWS CR's
	FirstV1Alpha3Release = "16.0.0"

	// FirstCiliumRelease is the first Cilium CNI GS release
	FirstCiliumRelease = "18.0.0-alpha1"

	// FirstOrgNamespaceRelease is the first GS release that creates Clusters in Org Namespaces by default
	FirstOrgNamespaceRelease = "16.0.0"

	// GiantSwarmLabelPart is the part of label keys that shows that they are protected giantswarm labels
	GiantSwarmLabelPart = "giantswarm.io"

	// GiantSwarmLabelPart is the part of label keys that shows that they are protected giantswarm labels
	ProviderTagLabelPart = "tag.provider.giantswarm.io"
)

const (
	// annotations should  taken from https://github.com/giantswarm/apiextensions/blob/master/pkg/annotation/aws.go
	// once the service is migrate to apiextensions v3
	AnnotationUpdateMaxBatchSize = "alpha.aws.giantswarm.io/update-max-batch-size"
	AnnotationUpdatePauseTime    = "alpha.aws.giantswarm.io/update-pause-time"

	AnnotationAlphaNodeTerminateUnhealthy = "alpha.node.giantswarm.io/terminate-unhealthy"
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

// ValidLabelAdmins returns the list of accounts used to manipulate labels
func ValidLabelAdmins() []string {
	return []string{
		"system:serviceaccount:giantswarm:api",
	}
}

// VersionLabels are the labels which are considered version labels
func VersionLabels() []string {
	return []string{label.Release, label.ClusterOperatorVersion}
}

// IsGiantSwarmLabel returns whether a label is considered a giantswarm label
func IsGiantSwarmLabel(label string) bool {
	return strings.Contains(label, GiantSwarmLabelPart)
}

// IsProviderTagLabel returns whether a label is considered a provider tag label
func IsProviderTagLabel(label string) bool {
	return strings.Contains(label, ProviderTagLabelPart)
}

// IsServicePriorityLabel returns whether a label is the service priority label
func IsServicePriorityLabel(l string) bool {
	return l == label.ServicePriority
}

// IsHAVersion returns whether a given releaseVersion supports HA Masters
func IsHAVersion(releaseVersion *semver.Version) bool {
	HAVersion, _ := semver.New(FirstHARelease)
	return releaseVersion.GE(*HAVersion)
}

// IsV1Alpha3Ready returns whether a given releaseVersion is a valid v1alpha3 release
func IsV1Alpha3Ready(releaseVersion *semver.Version) bool {
	V1Alpha3Version, _ := semver.New(FirstV1Alpha3Release)
	return releaseVersion.GE(*V1Alpha3Version)
}

// IsCiliumRelease returns whether a given releaseVersion is release with Cilium CNI
func IsCiliumRelease(releaseVersion *semver.Version) bool {
	V18Version, _ := semver.New(FirstCiliumRelease)
	return releaseVersion.GE(*V18Version)
}

// IsOrgNamespaceVersion returns whether a given releaseVersion creates clusters in org namespaces by default
func IsOrgNamespaceVersion(releaseVersion *semver.Version) bool {
	OrgNamespaceVersion, _ := semver.New(FirstOrgNamespaceRelease)
	return releaseVersion.GE(*OrgNamespaceVersion)
}

// IsCAPIVersion returns whether a given releaseVersion is using CAPI controllers
func IsCAPIVersion(releaseVersion *semver.Version) (bool, error) {
	CAPIVersion, err := semver.New(FirstCAPIRelease)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to get first CAPI release version")
	}
	return releaseVersion.GE(*CAPIVersion), nil
}

// IsVersionLabel returns whether a label is considered a version label
func IsVersionLabel(label string) bool {
	for _, l := range VersionLabels() {
		if l == label {
			return true
		}
	}
	return false
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
