// Package awsmachinedeployment intercepts write activity to AWSMachineDeployment objects.
package awsmachinedeployment

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
)

var (
	// If not specified otherwise, node pools should have 100% on-demand instances.
	defaultOnDemandPercentageAboveBaseCapacity int = 100
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

// Mutator for AWSMachineDeployment object.
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
	awsMachineDeploymentNewCR := &infrastructurev1alpha2.AWSMachineDeployment{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsMachineDeploymentNewCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse AWSMachineDeployment: %v", err)
	}
	patch, err = m.MutateOnDemandPercentage(*awsMachineDeploymentNewCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	patch, err = m.MutateReleaseVersion(*awsMachineDeploymentNewCR)
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

	// Parse incoming objects
	awsMachineDeploymentNewCR := &infrastructurev1alpha2.AWSMachineDeployment{}
	awsMachineDeploymentOldCR := &infrastructurev1alpha2.AWSMachineDeployment{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsMachineDeploymentNewCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse AWSMachineDeployment: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, awsMachineDeploymentOldCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse AWSMachineDeployment: %v", err)
	}
	patch, err = m.MutateOnDemandPercentage(*awsMachineDeploymentNewCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

// MutateOnDemandPercentage defaults the OnDemandPercentageAboveBaseCapacity.
func (m *Mutator) MutateOnDemandPercentage(awsMachineDeployment infrastructurev1alpha2.AWSMachineDeployment) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// Note: This will only work if the incoming CR has the .spec.provider.instanceDistribution
	// attribute defined. Otherwise the request to create/modify the CR will fail.
	if awsMachineDeployment.Spec.Provider.InstanceDistribution.OnDemandPercentageAboveBaseCapacity == nil {
		m.Log("level", "debug", "message", fmt.Sprintf("AWSMachineDeployment %s OnDemandPercentageAboveBaseCapacity is nil and will be set to default 100", awsMachineDeployment.ObjectMeta.Name))
		patch := mutator.PatchReplace("/spec/provider/instanceDistribution/onDemandPercentageAboveBaseCapacity", &defaultOnDemandPercentageAboveBaseCapacity)
		result = append(result, patch)
	}

	return result, nil
}
func (m *Mutator) MutateReleaseVersion(awsMachineDeployment infrastructurev1alpha2.AWSMachineDeployment) ([]mutator.PatchOperation, error) {
	return aws.MutateReleaseVersionLabel(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, &awsMachineDeployment)
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "awsmachinedeployment"
}
