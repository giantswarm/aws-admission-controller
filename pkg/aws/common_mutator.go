package aws

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
)

type Mutator struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

func MutateLabelFromAWSCluster(m *Mutator, meta metav1.Object, awsCluster infrastructurev1alpha2.AWSCluster, label string) ([]mutator.PatchOperation, error) {
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

func MutateLabelFromCluster(m *Mutator, meta metav1.Object, cluster capiv1alpha2.Cluster, label string) ([]mutator.PatchOperation, error) {
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

func MutateLabelFromRelease(m *Mutator, meta metav1.Object, release releasev1alpha1.Release, label string, component string) ([]mutator.PatchOperation, error) {
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

func FetchAWSCluster(m *Mutator, meta metav1.Object) (*infrastructurev1alpha2.AWSCluster, error) {
	var awsCluster infrastructurev1alpha2.AWSCluster
	var err error
	var fetch func() error

	namespace := meta.GetNamespace()
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}

	// Retrieve the Cluster ID.
	clusterID := key.Cluster(meta)
	if clusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "Object has no %s label, can't fetch AWSCluster.", label.Cluster)
	}

	// Fetch the AWSCluster CR
	{
		m.Logger.Log("level", "debug", "message", fmt.Sprintf("Fetching AWSCluster %s", clusterID))
		fetch = func() error {
			err := m.K8sClient.CtrlClient().Get(context.Background(), client.ObjectKey{Name: clusterID, Namespace: namespace}, &awsCluster)
			if IsNotFound(err) {
				return microerror.Maskf(notFoundError, "Looking for AWSCluster named %s but it was not found.", clusterID)
			} else if err != nil {
				return microerror.Mask(err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 10*time.Millisecond)
		err = backoff.Retry(fetch, b)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return &awsCluster, nil
}

func FetchCluster(m *Mutator, meta metav1.Object) (*capiv1alpha2.Cluster, error) {
	var cluster capiv1alpha2.Cluster
	var err error
	var fetch func() error

	namespace := meta.GetNamespace()
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	// Retrieve the Cluster ID.
	clusterID := key.Cluster(meta)
	if clusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "Object has no %s label, can't fetch cluster.", label.Cluster)
	}

	// Fetch the Cluster CR
	{
		m.Logger.Log("level", "debug", "message", fmt.Sprintf("Fetching Cluster %s", clusterID))
		fetch = func() error {
			err := m.K8sClient.CtrlClient().Get(context.Background(), client.ObjectKey{Name: clusterID, Namespace: namespace}, &cluster)
			if IsNotFound(err) {
				return microerror.Maskf(notFoundError, "Looking for Cluster named %s but it was not found.", clusterID)
			} else if err != nil {
				return microerror.Mask(err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 10*time.Millisecond)
		err = backoff.Retry(fetch, b)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return &cluster, nil
}

func FetchRelease(m *Mutator, version *semver.Version) (*releasev1alpha1.Release, error) {
	var releaseName string
	var release releasev1alpha1.Release
	var err error

	// Get release name
	{
		releaseName = version.String()
		if !strings.HasPrefix(releaseName, "v") {
			releaseName = fmt.Sprintf("v%s", releaseName)
		}
	}
	// Fetch the Release CR
	{
		m.Logger.Log("level", "debug", "message", fmt.Sprintf("Fetching Release %s", releaseName))
		err = m.K8sClient.CtrlClient().Get(context.Background(), client.ObjectKey{Name: releaseName, Namespace: metav1.NamespaceDefault}, &release)
		if IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "Looking for Release %s but it was not found.", releaseName)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return &release, nil
}

func GetNavailabilityZones(m *Mutator, n int, azs []string) []string {
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
