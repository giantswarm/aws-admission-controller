package g8scontrolplane

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/blang/semver"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	infrastructurev1alpha2scheme "github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned/scheme"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/reference"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws"
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
func (m *Mutator) MutateUpdate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	g8sControlPlaneNewCR := &infrastructurev1alpha2.G8sControlPlane{}
	g8sControlPlaneOldCR := &infrastructurev1alpha2.G8sControlPlane{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, g8sControlPlaneNewCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse g8scontrol plane: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, g8sControlPlaneOldCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse g8scontrol plane: %v", err)
	}

	releaseVersion, err := releaseVersion(g8sControlPlaneNewCR, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from G8sControlPlane")
	}

	// We try to fetch the AWSControlPlane belonging to the G8sControlPlane here.
	availabilityZones := 0
	awsControlPlane, err := m.fetchAWSControlPlane(*g8sControlPlaneNewCR)
	if IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the AWSControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		m.Log("level", "debug", "message", fmt.Sprintf("No AWSControlPlane %s could be found: %v", g8sControlPlaneNewCR.GetName(), err))
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		// This defaulting is only done when the awscontrolplane exists
		availabilityZones = len(awsControlPlane.Spec.AvailabilityZones)
		patch, err = m.MutateReplicaUpdate(*g8sControlPlaneNewCR, *g8sControlPlaneOldCR, *awsControlPlane)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	}

	patch, err = m.MutateReplicas(availabilityZones, *g8sControlPlaneNewCR, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

// MutateCreate is the function executed for every create webhook request.
func (m *Mutator) MutateCreate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	g8sControlPlaneCR := &infrastructurev1alpha2.G8sControlPlane{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, g8sControlPlaneCR); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse g8scontrol plane: %v", err)
	}

	patch, err = m.MutateReleaseVersion(*g8sControlPlaneCR)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	releaseVersion, err := releaseVersion(g8sControlPlaneCR, patch)
	if err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse release version from G8sControlPlane")
	}
	result = append(result, patch...)

	// We try to fetch the AWSControlPlane belonging to the G8sControlPlane here.
	availabilityZones := 0
	awsControlPlane, err := m.fetchAWSControlPlane(*g8sControlPlaneCR)
	if IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the AWSControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		m.Log("level", "debug", "message", fmt.Sprintf("No AWSControlPlane %s could be found: %v", g8sControlPlaneCR.GetName(), err))
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		// This defaulting is only done when the awscontrolplane exists
		availabilityZones = len(awsControlPlane.Spec.AvailabilityZones)
		patch, err = m.MutateInfraRef(*awsControlPlane, *g8sControlPlaneCR)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		result = append(result, patch...)
	}

	patch, err = m.MutateReplicas(availabilityZones, *g8sControlPlaneCR, releaseVersion)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}
func (m *Mutator) MutateReplicaUpdate(g8sControlPlaneNewCR infrastructurev1alpha2.G8sControlPlane, g8sControlPlaneOldCR infrastructurev1alpha2.G8sControlPlane, awsControlPlane infrastructurev1alpha2.AWSControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if its an update from single to HA master
	if !isUpdateFromSingleToHA(g8sControlPlaneNewCR, g8sControlPlaneOldCR, awsControlPlane) {
		return result, nil
	}
	// If the availability zones need to be updated from 1 to 3, we do it here
	update := func() error {
		ctx := context.Background()
		m.Log("level", "debug", "message", fmt.Sprintf("Updating AWSControlPlane AZs for HA %s", awsControlPlane.Name))
		awsControlPlane.Spec.AvailabilityZones = m.getHAavailabilityZones(awsControlPlane.Spec.AvailabilityZones[0], m.validAvailabilityZones)
		err := m.k8sClient.CtrlClient().Update(ctx, &awsControlPlane)
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	b := backoff.NewMaxRetries(3, 1*time.Second)
	err := backoff.Retry(update, b)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Mutator) MutateInfraRef(awsControlPlane infrastructurev1alpha2.AWSControlPlane, g8sControlPlane infrastructurev1alpha2.G8sControlPlane) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	if g8sControlPlane.Spec.InfrastructureRef.Name != "" && g8sControlPlane.Spec.InfrastructureRef.Namespace != "" {
		return result, nil
	}
	// We get the infrastructure reference
	infrastructureCRRef, err := reference.GetReference(infrastructurev1alpha2scheme.Scheme, &awsControlPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	m.Log("level", "debug", "message", fmt.Sprintf("Updating infrastructure reference to  %s", g8sControlPlane.Name))
	patch := mutator.PatchReplace("/spec/infrastructureRef", infrastructureCRRef)
	result = append(result, patch)
	return result, nil
}
func (m *Mutator) MutateReleaseVersion(g8sControlPlane infrastructurev1alpha2.G8sControlPlane) ([]mutator.PatchOperation, error) {
	return aws.MutateReleaseVersionLabel(&aws.Mutator{K8sClient: m.k8sClient, Logger: m.logger}, &g8sControlPlane)
}

func (m *Mutator) MutateReplicas(availabilityZones int, g8sControlPlane infrastructurev1alpha2.G8sControlPlane, releaseVersion *semver.Version) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	// We only need to manipulate if replicas are not set
	if g8sControlPlane.Spec.Replicas != 0 {
		return result, nil
	}
	var replicas int
	{
		replicas = aws.DefaultMasterReplicas
		// If there is an AWSControlPlane, the default replicas match the number of AZs
		if availabilityZones != 0 {
			replicas = availabilityZones
		}
		// For pre HA Masters, the replicas are 1 for a single master
		if !aws.IsHAVersion(releaseVersion) {
			replicas = 1
		}
	}
	// Trigger defaulting of the replicas
	m.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s Replicas are 0 and will be defaulted", g8sControlPlane.ObjectMeta.Name))
	patch := mutator.PatchReplace("/spec/replicas", replicas)
	result = append(result, patch)
	return result, nil
}

func (m *Mutator) fetchAWSControlPlane(g8sControlPlane infrastructurev1alpha2.G8sControlPlane) (*infrastructurev1alpha2.AWSControlPlane, error) {
	var awsControlPlane infrastructurev1alpha2.AWSControlPlane
	var err error
	var fetch func() error

	// Fetch the AWSControlPlane.
	{
		m.Log("level", "debug", "message", fmt.Sprintf("Fetching AWSControlPlane %s", g8sControlPlane.Name))
		fetch = func() error {
			ctx := context.Background()

			err = m.k8sClient.CtrlClient().Get(
				ctx,
				types.NamespacedName{Name: g8sControlPlane.GetName(), Namespace: g8sControlPlane.GetNamespace()},
				&awsControlPlane,
			)
			if err != nil {
				return microerror.Maskf(notFoundError, "failed to fetch AWSControlplane: %v", err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return &awsControlPlane, nil
}

func (m *Mutator) getHAavailabilityZones(firstAZ string, azs []string) []string {
	var randomAZs []string
	// Having 3 AZ's or more shuffle 3 HA masters in different AZ's
	if len(azs) >= 3 {
		var tempAZs []string
		for _, az := range azs {
			if firstAZ != az {
				tempAZs = append(tempAZs, az)
			}
		}
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(tempAZs), func(i, j int) { tempAZs[i], tempAZs[j] = tempAZs[j], tempAZs[i] })
		randomAZs = append(randomAZs, firstAZ, tempAZs[0], tempAZs[1])
		m.Log("level", "debug", "message", fmt.Sprintf("%d AZ's available, selected AZ's: %v", len(azs), randomAZs))

		return randomAZs

		// Having only 2 AZ available we shuffle 3 HA masters in 2 AZ's
	} else if len(azs) == 2 {
		var tempAZ string
		for _, az := range azs {
			if firstAZ != az {
				tempAZ = az
			}
		}
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(azs), func(i, j int) { azs[i], azs[j] = azs[j], azs[i] })
		randomAZs = append(randomAZs, firstAZ, tempAZ, azs[0])
		m.Log("level", "debug", "message", fmt.Sprintf("only %d AZ's available, random AZ's will be %v", len(azs), randomAZs))

		return randomAZs

		// Having only 1 AZ available we add 3 HA masters to this AZ
	} else {
		randomAZs = append(randomAZs, firstAZ, firstAZ, firstAZ)
		m.Log("level", "debug", "message", fmt.Sprintf("only %d AZ's available, using the same AZ %v", len(azs), randomAZs))

		return randomAZs
	}
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "g8scontrolplane"
}

func releaseVersion(cr *infrastructurev1alpha2.G8sControlPlane, patch []mutator.PatchOperation) (*semver.Version, error) {
	var version string
	var ok bool
	if len(patch) > 0 {
		if patch[0].Path == fmt.Sprintf("/metadata/labels/%s", aws.EscapeJSONPatchString(label.Release)) {
			version = patch[0].Value.(string)
		}
	} else {
		version, ok = cr.Labels[label.Release]
		if !ok {
			return nil, microerror.Maskf(parsingFailedError, "unable to get release version from G8sControlplane %s", cr.Name)
		}
	}
	return semver.New(version)
}

func isUpdateFromSingleToHA(g8sControlPlaneNewCR infrastructurev1alpha2.G8sControlPlane, g8sControlPlaneOldCR infrastructurev1alpha2.G8sControlPlane, awsControlPlane infrastructurev1alpha2.AWSControlPlane) bool {
	return g8sControlPlaneNewCR.Spec.Replicas == 3 && g8sControlPlaneOldCR.Spec.Replicas == 1 && len(awsControlPlane.Spec.AvailabilityZones) == 1
}
