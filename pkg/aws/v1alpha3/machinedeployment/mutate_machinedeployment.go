// Package machinedeployment intercepts write activity to MachineDeployment objects.
package machinedeployment

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
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

// Mutator for MachineDeployment object.
type Mutator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewMutator(config config.Config) (*Mutator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	mutator := &Mutator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
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
	machineDeployment := &capi.MachineDeployment{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, machineDeployment); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse MachineDeployment: %v", err)
	}
	capi, err := aws.IsCAPIRelease(machineDeployment)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if capi {
		return result, nil
	}

	patch, err = m.MutateReleaseVersion(*machineDeployment)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	releaseVersion, err := aws.ReleaseVersion(machineDeployment, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}

	patch = aws.MutateCAPILabel(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, machineDeployment)
	result = append(result, patch...)

	patch = m.MutateClusterName(*machineDeployment)
	result = append(result, patch...)

	patch = m.MutateTemplateClusterName(*machineDeployment)
	result = append(result, patch...)

	patch, err = m.MutateInfraRef(*machineDeployment, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

// MutateUpdate is the function executed for every update webhook request.
func (m *Mutator) MutateUpdate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	// Parse incoming object
	machineDeployment := &capi.MachineDeployment{}
	oldMachineDeployment := &capi.MachineDeployment{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, machineDeployment); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse Cluster: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, oldMachineDeployment); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse old Cluster: %v", err)
	}

	releaseVersion, err := aws.ReleaseVersion(machineDeployment, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}

	patch = aws.MutateCAPILabel(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, machineDeployment)
	result = append(result, patch...)

	patch = m.MutateClusterName(*machineDeployment)
	result = append(result, patch...)

	patch = m.MutateTemplateClusterName(*machineDeployment)
	result = append(result, patch...)

	patch, err = m.MutateInfraRef(*machineDeployment, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateReleaseVersion(machineDeployment capi.MachineDeployment) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	if key.Release(&machineDeployment) != "" && key.ClusterOperator(&machineDeployment) != "" {
		return result, nil
	}
	// Retrieve the `Cluster` CR related to this object.
	cluster, err := aws.FetchCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &machineDeployment)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the release label
	patch, err = aws.MutateLabelFromCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &machineDeployment, *cluster, label.Release)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	// mutate the operator label
	patch, err = aws.MutateLabelFromCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &machineDeployment, *cluster, label.ClusterOperatorVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateInfraRef(machineDeployment capi.MachineDeployment, releaseVersion *semver.Version) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	if machineDeployment.Spec.Template.Spec.InfrastructureRef.Name != "" && machineDeployment.Spec.Template.Spec.InfrastructureRef.Namespace != "" {
		return result, nil
	}
	namespace := machineDeployment.GetNamespace()
	if namespace == "" {
		namespace = v1.NamespaceDefault
	}

	var infrastructureCRRef v1.ObjectReference
	if aws.IsV1Alpha3Ready(releaseVersion) && machineDeployment.Spec.Template.Spec.InfrastructureRef.APIVersion == "infrastructure.giantswarm.io/v1alpha2" {
		infrastructureCRRef = v1.ObjectReference{
			APIVersion: "infrastructure.giantswarm.io/v1alpha3",
			Kind:       "AWSCluster",
			Name:       machineDeployment.GetName(),
			Namespace:  namespace,
		}
		m.Log("level", "debug", "message", fmt.Sprintf("Updating infrastructure reference to  %s", machineDeployment.Name))
		patch := mutator.PatchReplace("/spec/template/spec/infrastructureRef", &infrastructureCRRef)
		result = append(result, patch)
		return result, nil
	}
	return nil, nil
}

func (m *Mutator) MutateClusterName(machineDeployment capi.MachineDeployment) []mutator.PatchOperation {
	var result []mutator.PatchOperation
	if machineDeployment.Spec.ClusterName != "" {
		return result
	}

	m.Log("level", "debug", "message", fmt.Sprintf("Updating cluster name to %s", machineDeployment.Labels[label.Cluster]))
	patch := mutator.PatchReplace("/spec/clusterName", machineDeployment.Labels[label.Cluster])
	result = append(result, patch)
	return result
}

func (m *Mutator) MutateTemplateClusterName(machineDeployment capi.MachineDeployment) []mutator.PatchOperation {
	var result []mutator.PatchOperation
	if machineDeployment.Spec.Template.Spec.ClusterName != "" {
		return result
	}

	m.Log("level", "debug", "message", fmt.Sprintf("Updating template cluster name to %s", machineDeployment.Labels[label.Cluster]))
	patch := mutator.PatchReplace("/spec/template/spec/clusterName", machineDeployment.Labels[label.Cluster])
	result = append(result, patch)
	return result
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "machinedeployment"
}
