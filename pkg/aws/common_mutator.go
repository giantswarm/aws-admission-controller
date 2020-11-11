package aws

import (
	"context"
	"fmt"
	"strings"

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

		// Retrieve the `Cluster` CR related to this object.
		cluster := &capiv1alpha2.Cluster{}
		{
			err := m.K8sClient.CtrlClient().Get(context.Background(), client.ObjectKey{Name: clusterID, Namespace: meta.GetNamespace()}, cluster)
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

// Ensure the needed escapes are in place. See https://tools.ietf.org/html/rfc6901#section-3 .
func EscapeJSONPatchString(input string) string {
	input = strings.ReplaceAll(input, "~", "~0")
	input = strings.ReplaceAll(input, "/", "~1")

	return input
}
