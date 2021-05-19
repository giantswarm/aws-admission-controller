package aws

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
)

func MutateLabel(m *Handler, meta metav1.Object, label string, defaultValue string) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if meta.GetLabels()[label] != "" {
		return result, nil
	}

	m.Logger.Log("level", "debug", "message", fmt.Sprintf("Label %s is not set and will be defaulted to %s.",
		label,
		defaultValue))
	patch := mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label)), defaultValue)
	result = append(result, patch)

	return result, nil
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

	// Extract version from release
	value := GetReleaseComponentLabels(release)[component]
	if value == "" {
		return nil, microerror.Maskf(notFoundError, "Release %s did not specify version of %s.", release.GetName(), component)
	}
	if meta.GetLabels()[label] == value {
		return result, nil
	}
	m.Logger.Log("level", "debug", "message", fmt.Sprintf("Label %s will be defaulted to %s from Release %s.",
		label,
		value,
		release.GetName()))
	patch := mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label)), value)
	result = append(result, patch)

	return result, nil
}
