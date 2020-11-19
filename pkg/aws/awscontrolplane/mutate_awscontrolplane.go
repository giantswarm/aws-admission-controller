package awscontrolplane

import (
	"context"
	"fmt"
	"strings"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	infrastructurev1alpha2scheme "github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned/scheme"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/reference"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
)

const defaultnamespace = "default"

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
	releaseVersion, err := aws.ReleaseVersion(awsControlPlaneCR, patch)
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
	g8sControlPlane, err := m.fetchG8sControlPlane(*awsControlPlaneCR)
	if IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		m.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlaneCR.GetName(), err))
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		// This defaulting is only done when the awscontrolplane exists
		replicas = g8sControlPlane.Spec.Replicas
		patch, err = m.MutateInfraRef(*awsControlPlaneCR, *g8sControlPlane)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	}

	if aws.IsHAVersion(releaseVersion) {
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
	releaseVersion, err := aws.ReleaseVersion(awsControlPlaneCR, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from AWSControlPlane")
	}

	// We try to fetch the G8sControlPlane belonging to the AWSControlPlane here.
	replicas := 0
	g8sControlPlane, err := m.fetchG8sControlPlane(*awsControlPlaneCR)
	if IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		m.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlaneCR.GetName(), err))
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		// This defaulting is only done when the awscontrolplane exists
		replicas = g8sControlPlane.Spec.Replicas
	}

	if aws.IsHAVersion(releaseVersion) {
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

	awsCluster, err := aws.FetchAWSCluster(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane)
	if IsNotFound(err) {
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

func (m *Mutator) fetchG8sControlPlane(awsControlPlane infrastructurev1alpha2.AWSControlPlane) (*infrastructurev1alpha2.G8sControlPlane, error) {
	var g8sControlPlane infrastructurev1alpha2.G8sControlPlane
	var err error
	var fetch func() error

	namespace := awsControlPlane.GetNamespace()
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}

	// Fetch the G8sControlPlane.
	{
		m.Log("level", "debug", "message", fmt.Sprintf("Fetching G8sControlPlane %s", awsControlPlane.Name))
		fetch = func() error {
			ctx := context.Background()

			err = m.k8sClient.CtrlClient().Get(
				ctx,
				types.NamespacedName{Name: awsControlPlane.GetName(), Namespace: namespace},
				&g8sControlPlane,
			)
			if err != nil {
				return microerror.Maskf(notFoundError, "failed to fetch G8sControlplane: %v", err)
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
	return &g8sControlPlane, nil
}

func (m *Mutator) MutateAvailabilityZones(replicas int, awsControlPlaneCR infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if AZs are not set
	if awsControlPlaneCR.Spec.AvailabilityZones != nil {
		return result, nil
	}
	var numberOfAZs int
	{
		numberOfAZs = aws.DefaultMasterReplicas
		// If there is a G8sControlPlane, the default AZs match the replicas
		if replicas != 0 {
			numberOfAZs = replicas
		}
	}
	// Trigger defaulting of the master availability zones
	m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
	// We default the AZs
	defaultedAZs := aws.GetNavailabilityZones(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, numberOfAZs, m.validAvailabilityZones)
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

func (m *Mutator) MutateInfraRef(awsControlPlaneCR infrastructurev1alpha2.AWSControlPlane, g8sControlPlane infrastructurev1alpha2.G8sControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if the infraref is not set
	if g8sControlPlane.Spec.InfrastructureRef.Name != "" && g8sControlPlane.Spec.InfrastructureRef.Namespace != "" {
		return result, nil
	}

	update := func() error {
		ctx := context.Background()
		// If the infrastructure reference is not set, we do it here
		m.Log("level", "debug", "message", fmt.Sprintf("Updating infrastructure reference to  %s", awsControlPlaneCR.Name))
		infrastructureCRRef, err := reference.GetReference(infrastructurev1alpha2scheme.Scheme, &awsControlPlaneCR)
		if err != nil {
			return microerror.Mask(err)
		}
		if infrastructureCRRef.Namespace == "" {
			infrastructureCRRef.Namespace = defaultnamespace
		}
		g8sControlPlane.Spec.InfrastructureRef = *infrastructureCRRef
		err = m.k8sClient.CtrlClient().Update(ctx, &g8sControlPlane)
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	b := backoff.NewMaxRetries(3, 10*time.Millisecond)
	err := backoff.Retry(update, b)
	if err != nil {
		return nil, err
	}
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
	patch := mutator.PatchAdd("/spec/instanceType", aws.DefaultMasterInstanceType)
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
	awsCluster, err := aws.FetchAWSCluster(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the operator label
	patch, err = aws.MutateLabelFromAWSCluster(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane, *awsCluster, label.AWSOperatorVersion)
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
	cluster, err := aws.FetchCluster(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// mutate the release label
	patch, err = aws.MutateLabelFromCluster(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, &awsControlPlane, *cluster, label.Release)
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
