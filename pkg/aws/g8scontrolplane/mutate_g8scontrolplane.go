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
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/reference"

	"github.com/giantswarm/aws-admission-controller/config"
	"github.com/giantswarm/aws-admission-controller/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
)

type Mutator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	validAvailabilityZones []string
}

func NewMutator(config config.Config) (*Mutator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(aws.InvalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(aws.InvalidConfigError, "%T.Logger must not be empty", config)
	}

	var availabilityZones []string = strings.Split(config.AvailabilityZones, ",")
	mutator := &Mutator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		validAvailabilityZones: availabilityZones,
	}

	return mutator, nil
}

func (m *Mutator) Mutate(request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		return result, nil
	}

	g8sControlPlaneNewCR := &infrastructurev1alpha2.G8sControlPlane{}
	g8sControlPlaneOldCR := &infrastructurev1alpha2.G8sControlPlane{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, g8sControlPlaneNewCR); err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse g8scontrol plane: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, g8sControlPlaneOldCR); err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse g8scontrol plane: %v", err)
	}
	releaseVersion, err := releaseVersion(g8sControlPlaneNewCR)
	if err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse release version from AWSControlPlane")
	}
	namespace := g8sControlPlaneNewCR.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	var replicas int

	// We only need to manipulate if replicas are not set or if its an update from single to HA master or on create
	if g8sControlPlaneNewCR.Spec.Replicas != 0 && !isUpdateFromSingleToHA(g8sControlPlaneNewCR, g8sControlPlaneOldCR) && request.Operation != aws.CreateOperation {
		return result, nil
	}
	infrastructureCRRef := &v1.ObjectReference{}
	// We need to fetch the AWSControlPlane in case AZs need to be defaulted or the g8scontrolplane is just created
	if aws.IsHAVersion(releaseVersion) || request.Operation == aws.CreateOperation {
		replicas = aws.DefaultMasterReplicas
		update := func() error {
			ctx := context.Background()

			// We fetch the AWSControlPlane CR.
			awsControlPlane := &infrastructurev1alpha2.AWSControlPlane{}
			{
				m.Log("level", "debug", "message", fmt.Sprintf("Fetching AWSControlPlane %s", g8sControlPlaneNewCR.Name))
				err := m.k8sClient.CtrlClient().Get(ctx,
					types.NamespacedName{Name: g8sControlPlaneNewCR.GetName(),
						Namespace: namespace},
					awsControlPlane)
				if err != nil {
					return microerror.Maskf(aws.NotFoundError, "failed to fetch AWSControlplane: %v", err)
				}
			}
			// If there is an AWSControlPlane, the default replicas match the number of AZs
			replicas = len(awsControlPlane.Spec.AvailabilityZones)
			// If there is an AWSControlplane, we get its infrastructure reference
			infrastructureCRRef, err = reference.GetReference(infrastructurev1alpha2scheme.Scheme, awsControlPlane)
			if err != nil {
				return microerror.Mask(err)
			}

			// If the availability zones need to be updated from 1 to 3, we do it here
			{
				if aws.IsHAVersion(releaseVersion) && isUpdateFromSingleToHA(g8sControlPlaneNewCR, g8sControlPlaneOldCR) && len(awsControlPlane.Spec.AvailabilityZones) == 1 {
					m.Log("level", "debug", "message", fmt.Sprintf("Updating AWSControlPlane AZs for HA %s", awsControlPlane.Name))
					awsControlPlane.Spec.AvailabilityZones = m.getHAavailabilityZones(awsControlPlane.Spec.AvailabilityZones[0], m.validAvailabilityZones)
					err := m.k8sClient.CtrlClient().Update(ctx, awsControlPlane)
					if err != nil {
						return microerror.Mask(err)
					}
				}
				return nil
			}
		}
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err := backoff.Retry(update, b)
		// Note that while we do log the error, we don't fail if the AWSControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		if aws.IsNotFound(err) {
			m.Log("level", "debug", "message", fmt.Sprintf("No AWSControlPlane %s could be found: %v", g8sControlPlaneNewCR.Name, err))
		} else if err != nil {
			return nil, err
		}
	}
	// For pre HA Masters, the replicas are 1 for a single master
	if !aws.IsHAVersion(releaseVersion) {
		replicas = 1
	}
	// Trigger defaulting of the replicas
	if g8sControlPlaneNewCR.Spec.Replicas == 0 {
		m.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s Replicas are 0 and will be defaulted", g8sControlPlaneNewCR.ObjectMeta.Name))
		patch := mutator.PatchReplace("/spec/replicas", replicas)
		result = append(result, patch)
	}
	// If the infrastructure reference is not set, we do it here
	if request.Operation == aws.CreateOperation && g8sControlPlaneNewCR.Spec.InfrastructureRef.Name != infrastructureCRRef.Name {
		m.Log("level", "debug", "message", fmt.Sprintf("Updating infrastructure reference to  %s", g8sControlPlaneNewCR.Name))
		patch := mutator.PatchReplace("/spec/infrastructureRef", infrastructureCRRef)
		result = append(result, patch)
	}

	return result, nil
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

func releaseVersion(cr *infrastructurev1alpha2.G8sControlPlane) (*semver.Version, error) {
	version, ok := cr.Labels[label.ReleaseVersion]
	if !ok {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to get release version from G8sControlplane %s", cr.Name)
	}

	return semver.New(version)
}

func isUpdateFromSingleToHA(g8sControlPlaneNewCR *infrastructurev1alpha2.G8sControlPlane, g8sControlPlaneOldCR *infrastructurev1alpha2.G8sControlPlane) bool {
	return g8sControlPlaneNewCR.Spec.Replicas == 3 && g8sControlPlaneOldCR.Spec.Replicas == 1
}
