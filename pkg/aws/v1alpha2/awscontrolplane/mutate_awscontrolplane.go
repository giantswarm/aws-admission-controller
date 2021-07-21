package awscontrolplane

import (
	"fmt"
	"strings"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	awsv1alpha2 "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
)

type Mutator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	validAvailabilityZones []string
}

func NewMutator(config config.Config) (*Mutator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var availabilityZones []string = strings.Split(config.AvailabilityZones, ",")
	mutator := &Mutator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		validAvailabilityZones: availabilityZones,
	}

	return mutator, nil
}

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

	awsControlPlaneCR := &infrastructurev1alpha2.AWSControlPlane{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsControlPlaneCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse awscontrol plane: %v", err)
	}

	patch, err = m.MutateReleaseVersion(*awsControlPlaneCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	releaseVersion, err := awsv1alpha2.ReleaseVersion(awsControlPlaneCR, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from AWSControlPlane")
	}
	result = append(result, patch...)

	patch, err = m.MutateOperatorVersion(*awsControlPlaneCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	// We try to fetch the G8sControlPlane belonging to the AWSControlPlane here.
	replicas := 0
	g8sControlPlane, err := awsv1alpha2.FetchG8sControlPlane(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, awsControlPlaneCR)
	if awsv1alpha2.IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		m.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlaneCR.GetName(), err))
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		// This defaulting is only done when the awscontrolplane exists
		replicas = g8sControlPlane.Spec.Replicas
	}

	if awsv1alpha2.IsHAVersion(releaseVersion) {
		patch, err = m.MutateInstanceType(*awsControlPlaneCR)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)

		patch, err = m.MutateAvailabilityZones(replicas, *awsControlPlaneCR)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	} else {
		patch, err = m.MutatePreHA(*awsControlPlaneCR)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	}

	return result, nil
}

// MutateUpdate is the function executed for every update webhook request.
func (m *Mutator) MutateUpdate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	awsControlPlaneCR := &infrastructurev1alpha2.AWSControlPlane{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsControlPlaneCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	releaseVersion, err := awsv1alpha2.ReleaseVersion(awsControlPlaneCR, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from AWSControlPlane")
	}

	// We try to fetch the G8sControlPlane belonging to the AWSControlPlane here.
	replicas := 0
	g8sControlPlane, err := awsv1alpha2.FetchG8sControlPlane(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, awsControlPlaneCR)
	if awsv1alpha2.IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		m.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlaneCR.GetName(), err))
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		// This defaulting is only done when the awscontrolplane exists
		replicas = g8sControlPlane.Spec.Replicas
	}

	if awsv1alpha2.IsHAVersion(releaseVersion) {
		patch, err = m.MutateInstanceType(*awsControlPlaneCR)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)

		patch, err = m.MutateAvailabilityZones(replicas, *awsControlPlaneCR)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	} else {
		patch, err = m.MutatePreHA(*awsControlPlaneCR)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	}

	return result, nil
}

// MutatePreHA is there to mutate the master instance attributes from the AWSCluster CR in legacy versions.
// This can be deprecated once no versions < 11.4.0 are in use anymore
func (m *Mutator) MutatePreHA(awsControlPlane infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	awsCluster, err := awsv1alpha2.FetchAWSCluster(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane)
	if awsv1alpha2.IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the AWSCluster doesn't exist yet. That is okay because the order of CR creation can vary.
		// In this case we simply default as usual with one AZ.
		m.Log("level", "debug", "message", fmt.Sprintf("No AWSCluster %s could be found: %v", awsControlPlane.GetName(), err))
		patch, err = m.MutateAvailabilityZones(1, awsControlPlane)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)

		patch, err = m.MutateInstanceType(awsControlPlane)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		patch, err = m.MutateAvailabilityZonesPreHA([]string{awsCluster.Spec.Provider.Master.AvailabilityZone}, awsControlPlane)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)

		patch, err = m.MutateInstanceTypePreHA(awsCluster.Spec.Provider.Master.InstanceType, awsControlPlane)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	}
	return result, nil
}

func (m *Mutator) MutateAvailabilityZones(replicas int, awsControlPlaneCR infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if AZs are not set
	if awsControlPlaneCR.Spec.AvailabilityZones != nil {
		return result, nil
	}
	var numberOfAZs int
	{
		numberOfAZs = awsv1alpha2.DefaultMasterReplicas
		// If there is a G8sControlPlane, the default AZs match the replicas
		if replicas != 0 {
			numberOfAZs = replicas
		}
	}
	// Trigger defaulting of the master availability zones
	m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
	// We default the AZs
	defaultedAZs := awsv1alpha2.GetNavailabilityZones(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, numberOfAZs, m.validAvailabilityZones)
	patch := mutator.PatchAdd("/spec/availabilityZones", defaultedAZs)
	result = append(result, patch)
	return result, nil
}

func (m *Mutator) MutateAvailabilityZonesPreHA(availabilityZone []string, awsControlPlaneCR infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if AZs are not set
	if awsControlPlaneCR.Spec.AvailabilityZones != nil {
		return result, nil
	}
	m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
	patch := mutator.PatchAdd("/spec/availabilityZones", availabilityZone)
	result = append(result, patch)
	return result, nil
}

func (m *Mutator) MutateInstanceType(awsControlPlaneCR infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if instance type is not set
	if awsControlPlaneCR.Spec.InstanceType != "" {
		return result, nil
	}
	// Trigger defaulting of the master instance type
	m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s InstanceType is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
	patch := mutator.PatchAdd("/spec/instanceType", awsv1alpha2.DefaultMasterInstanceType)
	result = append(result, patch)
	return result, nil
}

func (m *Mutator) MutateInstanceTypePreHA(instanceType string, awsControlPlaneCR infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if instance type is not set
	if awsControlPlaneCR.Spec.InstanceType != "" {
		return result, nil
	}
	// Trigger defaulting of the master instance type
	m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s InstanceType is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
	patch := mutator.PatchAdd("/spec/instanceType", instanceType)
	result = append(result, patch)
	return result, nil
}

func (m *Mutator) MutateOperatorVersion(awsControlPlane infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	if key.AWSOperator(&awsControlPlane) != "" {
		return result, nil
	}
	// Retrieve the `AWSCluster` CR related to this object.
	awsCluster, err := awsv1alpha2.FetchAWSCluster(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the operator label
	patch, err = awsv1alpha2.MutateLabelFromAWSCluster(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane, *awsCluster, label.AWSOperatorVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

func (m *Mutator) MutateReleaseVersion(awsControlPlane infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	if key.Release(&awsControlPlane) != "" {
		return result, nil
	}
	// Retrieve the `Cluster` CR related to this object.
	cluster, err := awsv1alpha2.FetchCluster(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the release label
	patch, err = awsv1alpha2.MutateLabelFromCluster(&awsv1alpha2.Handler{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane, *cluster, label.Release)
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
	return "awscontrolplane"
}
