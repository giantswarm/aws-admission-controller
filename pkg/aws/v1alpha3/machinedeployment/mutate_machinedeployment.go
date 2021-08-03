// Package machinedeployment intercepts write activity to MachineDeployment objects.
package machinedeployment

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/aws-admission-controller/v3/config"
	aws "github.com/giantswarm/aws-admission-controller/v3/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/mutator"
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
	machineDeployment := &capiv1alpha3.MachineDeployment{}
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
	cluster := &capiv1alpha3.MachineDeployment{}
	oldCluster := &capiv1alpha3.MachineDeployment{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, cluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse Cluster: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, oldCluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse old Cluster: %v", err)
	}

	releaseVersion, err := aws.ReleaseVersion(cluster, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}
	patch, err = m.MutateInfraRef(*cluster, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateReleaseVersion(machineDeployment capiv1alpha3.MachineDeployment) ([]mutator.PatchOperation, error) {
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

func (m *Mutator) MutateInfraRef(machineDeployment capiv1alpha3.MachineDeployment, releaseVersion *semver.Version) ([]mutator.PatchOperation, error) {
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
		patch := mutator.PatchReplace("/spec/infrastructureRef", &infrastructureCRRef)
		result = append(result, patch)
		return result, nil
	}
	return nil, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "machinedeployment"
}
