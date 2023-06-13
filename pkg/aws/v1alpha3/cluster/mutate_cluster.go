// Package cluster intercepts write activity to Cluster objects.
package cluster

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/aws-admission-controller/v4/config"
	aws "github.com/giantswarm/aws-admission-controller/v4/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/mutator"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

// Mutator for Cluster object.
type Mutator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	defaultAWSCNIPodCidr string
	defaultCiliumPodCidr string
}

func NewMutator(config config.Config) (*Mutator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.CiliumDefaultPodCidr == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CiliumDefaultPodCidr must not be empty", config)
	}

	mutator := &Mutator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		defaultAWSCNIPodCidr: fmt.Sprintf("%s/%s", config.PodSubnet, config.PodCIDR),
		defaultCiliumPodCidr: config.CiliumDefaultPodCidr,
	}

	return mutator, nil
}

// Mutate is the function executed for every matching webhook request.
func (m *Mutator) Mutate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		return result, nil
	}
	if request.Operation == admissionv1.Create {
		return m.MutateCreate(request)
	}
	if request.Operation == admissionv1.Update {
		return m.MutateUpdate(request)
	}
	return result, nil
}

// MutateCreate is the function executed for every create webhook request.
func (m *Mutator) MutateCreate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	// Parse incoming object
	cluster := &capi.Cluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, cluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse Cluster: %v", err)
	}

	capi, err := aws.IsCAPIRelease(cluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if capi {
		return result, nil
	}

	patch, err = m.MutateReleaseVersion(*cluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	releaseVersion, err := aws.ReleaseVersion(cluster, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}
	result = append(result, patch...)

	patch, err = m.MutateOperatorVersion(*cluster, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	patch = aws.MutateCAPILabel(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, cluster)
	result = append(result, patch...)

	patch, err = m.MutateInfraRef(*cluster, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	patch = m.DefaultCiliumIpamMode(*cluster, releaseVersion)
	result = append(result, patch...)

	return result, nil
}

// MutateUpdate is the function executed for every update webhook request.
func (m *Mutator) MutateUpdate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	// Parse incoming object
	cluster := &capi.Cluster{}
	oldCluster := &capi.Cluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, cluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse Cluster: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, oldCluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse old Cluster: %v", err)
	}

	capi, err := aws.IsCAPIRelease(cluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if capi {
		return result, nil
	}

	patch, err = m.MutateReleaseUpdate(*cluster, *oldCluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	releaseVersion, err := aws.ReleaseVersion(cluster, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}

	patch = aws.MutateCAPILabel(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, cluster)
	result = append(result, patch...)

	oldReleaseVersion := semver.MustParse(oldCluster.Labels[label.Release])

	patch, err = m.MutateInfraRef(*cluster, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	patch, err = m.DefaultCiliumCidrOnV18Upgrade(*cluster, &oldReleaseVersion, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateOperatorVersion(cluster capi.Cluster, releaseVersion *semver.Version) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	if key.ClusterOperator(&cluster) != "" {
		return result, nil
	}
	// Retrieve the `Release` CR.
	release, err := aws.FetchRelease(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the operator label
	patch, err = aws.MutateLabelFromRelease(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &cluster, *release, label.ClusterOperatorVersion, "cluster-operator")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateReleaseVersion(cluster capi.Cluster) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var err error

	if key.Release(&cluster) != "" {
		return result, nil
	}
	// Find the newest active release.
	newestRelease, err := aws.FetchNewestReleaseVersion(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the release label
	m.Log("level", "debug", "message", fmt.Sprintf("Label %s is not set and will be defaulted to newest version %s.",
		label.Release,
		newestRelease.String()))
	patch := mutator.PatchAdd(fmt.Sprintf("/metadata/labels/%s", aws.EscapeJSONPatchString(label.Release)), newestRelease.String())
	result = append(result, patch)

	return result, nil
}

func (m *Mutator) MutateReleaseUpdate(cluster capi.Cluster, oldCluster capi.Cluster) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	if key.Release(&cluster) == key.Release(&oldCluster) {
		return result, nil
	}
	// Retrieve the `Release` CR.
	releaseVersion, err := aws.ReleaseVersion(&cluster, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}
	release, err := aws.FetchRelease(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the operator label
	patch, err = aws.MutateLabelFromRelease(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &cluster, *release, label.ClusterOperatorVersion, "cluster-operator")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "cluster"
}

func (m *Mutator) MutateInfraRef(cluster capi.Cluster, releaseVersion *semver.Version) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	if cluster.Spec.InfrastructureRef.Name != "" && cluster.Spec.InfrastructureRef.Namespace != "" {
		return result, nil
	}
	namespace := cluster.GetNamespace()
	if namespace == "" {
		namespace = v1.NamespaceDefault
	}

	var infrastructureCRRef v1.ObjectReference
	if aws.IsV1Alpha3Ready(releaseVersion) && cluster.Spec.InfrastructureRef.APIVersion == "infrastructure.giantswarm.io/v1alpha2" {
		infrastructureCRRef = v1.ObjectReference{
			APIVersion: "infrastructure.giantswarm.io/v1alpha3",
			Kind:       "AWSCluster",
			Name:       cluster.GetName(),
			Namespace:  namespace,
		}
		m.Log("level", "debug", "message", fmt.Sprintf("Updating infrastructure reference to  %s", cluster.Name))
		patch := mutator.PatchReplace("/spec/infrastructureRef", &infrastructureCRRef)
		result = append(result, patch)
		return result, nil
	}
	return nil, nil
}

func (m *Mutator) DefaultCiliumCidrOnV18Upgrade(cluster capi.Cluster, currentRelease *semver.Version, targetRelease *semver.Version) ([]mutator.PatchOperation, error) {
	if aws.IsPreCiliumRelease(currentRelease) && aws.IsPreCiliumRelease(targetRelease) || aws.IsCiliumRelease(currentRelease) && aws.IsCiliumRelease(targetRelease) {
		return nil, nil
	}

	annotations := cluster.Annotations

	// Default the cilium pod cidr annotation if it's not set already.
	if _, ok := cluster.Annotations[annotation.CiliumPodCidr]; !ok {
		// We only default the cilium CIDR if this cluster:
		// - is not using networkpools
		// - is using the default pod cidr

		// Retrieve the `AWSCluster` CR related to this object.
		awsCluster, err := aws.FetchAWSCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &cluster)
		if apierrors.IsNotFound(err) {
			// No AWS cluster exists, can't provide a default.
			m.Log("level", "debug", "message", "AWSCluster not found, can't default cilium cidr")
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if awsCluster.Spec.Provider.Nodes.NetworkPool != "" {
			// Networkpool in use, can't provide a sane default.
			m.Log("level", "debug", "message", "Networkpool is set, can't default cilium cidr")
			return nil, nil
		}

		if awsCluster.Spec.Provider.Pods.CIDRBlock != m.defaultAWSCNIPodCidr {
			// Non default pod cidr, can't provide a sane default.
			m.Log("level", "debug", "message", "Using not default cidr block, can't default cilium cidr")
			return nil, nil
		}

		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[annotation.CiliumPodCidr] = m.defaultCiliumPodCidr
	}

	// When moving to cilium, we don't want to have cilium be kube-proxy implementation from the beginning as
	// this will cause a downtime in the customer workloads. This annotation disables the feature.
	// AWS-operator will enable the feature again once all nodes are rolled.
	annotations[annotation.CiliumForceDisableKubeProxyAnnotation] = "true"

	var result []mutator.PatchOperation
	patch := mutator.PatchAdd("/metadata/annotations", annotations)
	result = append(result, patch)
	return result, nil
}

func (m *Mutator) DefaultCiliumIpamMode(cluster capi.Cluster, release *semver.Version) []mutator.PatchOperation {
	if aws.IsPreCiliumRelease(release) {
		// No need for the annotation before cilium.
		return nil
	}

	annotations := cluster.Annotations
	if annotations == nil {
		annotations = make(map[string]string, 0)
	}

	_, found := annotations[annotation.CiliumIpamModeAnnotation]
	if found {
		// Annotation already present.
		return nil
	}

	annotations[annotation.CiliumIpamModeAnnotation] = annotation.CiliumIpamModeKubernetes

	var result []mutator.PatchOperation
	patch := mutator.PatchAdd("/metadata/annotations", annotations)
	result = append(result, patch)
	return result
}
