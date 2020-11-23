package aws

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
)

type Handler struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

func MutateLabelFromAWSCluster(m *Handler, meta metav1.Object, awsCluster infrastructurev1alpha2.AWSCluster, label string) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if meta.GetLabels()[label] != "" {
		return result, nil
	}

	// Extract release from Cluster.
	value := awsCluster.GetLabels()[label]
	if value == "" {
		return nil, microerror.Maskf(notFoundError, "AWSCluster %s did not have the label %s set.", awsCluster.GetName(), label)
	}
	m.Logger.Log("level", "debug", "message", fmt.Sprintf("Label %s is not set and will be defaulted to %s from AWSCluster %s.",
		label,
		value,
		awsCluster.GetName()))
	patch := mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label)), value)
	result = append(result, patch)

	return result, nil
}

func MutateLabelFromCluster(m *Handler, meta metav1.Object, cluster capiv1alpha2.Cluster, label string) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if meta.GetLabels()[label] != "" {
		return result, nil
	}

	// Extract release from Cluster.
	value := cluster.GetLabels()[label]
	if value == "" {
		return nil, microerror.Maskf(notFoundError, "Cluster %s did not have the label %s set.", cluster.GetName(), label)
	}
	m.Logger.Log("level", "debug", "message", fmt.Sprintf("Label %s is not set and will be defaulted to %s from Cluster %s.",
		label,
		value,
		cluster.GetName()))
	patch := mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label)), value)
	result = append(result, patch)

	return result, nil
}

func MutateLabelFromRelease(m *Handler, meta metav1.Object, release releasev1alpha1.Release, label string, component string) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if meta.GetLabels()[label] != "" {
		return result, nil
	}
	// Extract version from release
	value := GetReleaseComponentLabels(release)[component]
	if value == "" {
		return nil, microerror.Maskf(notFoundError, "Release %s did not specify version of %s.", release.GetName(), component)
	}
	m.Logger.Log("level", "debug", "message", fmt.Sprintf("Label %s is not set and will be defaulted to %s from Release %s.",
		label,
		value,
		release.GetName()))
	patch := mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label)), value)
	result = append(result, patch)

	return result, nil
}
func GetReleaseComponentLabels(release releasev1alpha1.Release) map[string]string {
	components := map[string]string{}
	for _, component := range release.Spec.Components {
		components[component.Name] = component.Version
	}
	return components
}

func GetNavailabilityZones(m *Handler, n int, azs []string) []string {
	randomAZs := azs
	// In case there are not enough distinct AZs, we repeat them
	for len(randomAZs) < n {
		randomAZs = append(randomAZs, azs...)
	}
	// We shuffle the AZs, pick the first n and sort them alphabetically
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(randomAZs), func(i, j int) { randomAZs[i], randomAZs[j] = randomAZs[j], randomAZs[i] })
	randomAZs = randomAZs[:n]
	sort.Strings(randomAZs)
	m.Logger.Log("level", "debug", "message", fmt.Sprintf("available AZ's: %v, selected AZ's: %v", azs, randomAZs))

	return randomAZs
}

func ReleaseVersion(meta metav1.Object, patch []mutator.PatchOperation) (*semver.Version, error) {
	var version string
	var ok bool
	if len(patch) > 0 {
		if patch[0].Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.Release)) {
			version = patch[0].Value.(string)
		}
	} else {
		version, ok = meta.GetLabels()[label.Release]
		if !ok {
			return nil, microerror.Maskf(parsingFailedError, "unable to get release version from Object %s", meta.GetName())
		}
	}
	return semver.New(version)
}

// Ensure the needed escapes are in place. See https://tools.ietf.org/html/rfc6901#section-3 .
func EscapeJSONPatchString(input string) string {
	input = strings.ReplaceAll(input, "~", "~0")
	input = strings.ReplaceAll(input, "/", "~1")

	return input
}
