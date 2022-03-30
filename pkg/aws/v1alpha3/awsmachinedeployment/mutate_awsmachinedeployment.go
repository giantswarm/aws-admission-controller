// package awsmachinedeployment intercepts write activity to AWSMachineDeployment objects.
package awsmachinedeployment

import (
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v3/config"
	aws "github.com/giantswarm/aws-admission-controller/v3/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/mutator"
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
	awsMachineDeploymentNewCR := &infrastructurev1alpha3.AWSMachineDeployment{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsMachineDeploymentNewCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse AWSMachineDeployment: %v", err)
	}
	patch, err = m.MutateAvailabilityZones(*awsMachineDeploymentNewCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	patch, err = m.MutateOnDemandPercentage(*awsMachineDeploymentNewCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	patch = aws.MutateCAPILabel(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, awsMachineDeploymentNewCR)
	result = append(result, patch...)

	patch, err = m.MutateReleaseVersion(*awsMachineDeploymentNewCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	patch, err = m.MutateOperatorVersion(*awsMachineDeploymentNewCR)
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
	awsMachineDeploymentNewCR := &infrastructurev1alpha3.AWSMachineDeployment{}
	awsMachineDeploymentOldCR := &infrastructurev1alpha3.AWSMachineDeployment{}
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

	patch = aws.MutateCAPILabel(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, awsMachineDeploymentNewCR)
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateAvailabilityZones(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if AZs are not set
	if len(awsMachineDeployment.Spec.Provider.AvailabilityZones) != 0 {
		return result, nil
	}

	// Retrieve the `AWSControlPlane` CR related to this object.
	awsControlPlane, err := aws.FetchAWSControlPlane(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsMachineDeployment)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	// return early if no AZs are given
	if len(awsControlPlane.Spec.AvailabilityZones) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "No availability zones assigned in AWSControlPlane %s.", awsControlPlane.GetName())
	}
	// Trigger defaulting of the worker availability zones
	m.Log("level", "debug", "message", fmt.Sprintf("AWSMachineDeployment %s AvailabilityZones are not set and will be defaulted", awsMachineDeployment.ObjectMeta.Name))
	// We default the AZs
	defaultedAZs := aws.GetNavailabilityZones(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, aws.DefaultNodePoolAZs, awsControlPlane.Spec.AvailabilityZones)
	patch := mutator.PatchAdd("/spec/provider/availabilityZones", defaultedAZs)
	result = append(result, patch)
	return result, nil
}

// MutateOnDemandPercentage defaults the OnDemandPercentageAboveBaseCapacity.
func (m *Mutator) MutateOnDemandPercentage(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) ([]mutator.PatchOperation, error) {
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

func (m *Mutator) MutateOperatorVersion(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	if key.AWSOperator(&awsMachineDeployment) != "" {
		return result, nil
	}
	// Retrieve the `AWSCluster` CR related to this object.
	awsCluster, err := aws.FetchAWSCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsMachineDeployment)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the operator label
	patch, err = aws.MutateLabelFromAWSCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsMachineDeployment, *awsCluster, label.AWSOperatorVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateReleaseVersion(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	if key.Release(&awsMachineDeployment) != "" {
		return result, nil
	}
	// Retrieve the `Cluster` CR related to this object.
	cluster, err := aws.FetchCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsMachineDeployment)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the release label
	patch, err = aws.MutateLabelFromCluster(&aws.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsMachineDeployment, *cluster, label.Release)
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
	return "awsmachinedeployment"
}
