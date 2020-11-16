package aws

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
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

func MutateReleaseVersionLabel(m *Mutator, meta metav1.Object) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if key.Release(meta) != "" {
		return result, nil
	}
	// If the release version label is not set, we default it here
	{
		// Retrieve the Cluster ID.
		clusterID := key.Cluster(meta)
		if clusterID == "" {
			return nil, microerror.Maskf(invalidConfigError, "Object has no %s label, can't detect release version.", label.Cluster)
		}

		namespace := meta.GetNamespace()
		if namespace == "" {
			namespace = metav1.NamespaceDefault
		}

		// Retrieve the `Cluster` CR related to this object.
		cluster := &capiv1alpha2.Cluster{}
		{
			err := m.K8sClient.CtrlClient().Get(context.Background(), client.ObjectKey{Name: clusterID, Namespace: namespace}, cluster)
			if IsNotFound(err) {
				return nil, microerror.Maskf(notFoundError, "Looking for Cluster named %s but it was not found.", clusterID)
			} else if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		// Extract release from Cluster.
		release := key.Release(cluster)
		if release == "" {
			return nil, microerror.Maskf(notFoundError, "Cluster %s did not have a release label set.", clusterID)
		}
		m.Logger.Log("level", "debug", "message", fmt.Sprintf("Release label is not set and will be defaulted to %s from Cluster %s.",
			release,
			cluster.GetName()))
		patch := mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.Release)), release)
		result = append(result, patch)
	}

	return result, nil
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
