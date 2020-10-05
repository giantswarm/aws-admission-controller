package awscontrolplane

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
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

	awsControlPlaneCR := &infrastructurev1alpha2.AWSControlPlane{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsControlPlaneCR); err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	releaseVersion, err := releaseVersion(awsControlPlaneCR)
	if err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse release version from AWSControlPlane")
	}
	namespace := awsControlPlaneCR.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}
	var numberOfAZs int

	// We only need to manipulate if attributes are not set or it's a create operation
	if awsControlPlaneCR.Spec.AvailabilityZones != nil && awsControlPlaneCR.Spec.InstanceType != "" && request.Operation != aws.CreateOperation {
		return result, nil
	}
	// We need to fetch the G8sControlPlane in case AZs need to be defaulted or the awscontrolplane is just created
	if (aws.IsHAVersion(releaseVersion) && awsControlPlaneCR.Spec.AvailabilityZones == nil) || request.Operation == aws.CreateOperation {
		numberOfAZs = aws.DefaultMasterReplicas
		fetch := func() error {
			ctx := context.Background()

			// We try to fetch the G8sControlPlane CR.
			g8sControlPlane := &infrastructurev1alpha2.G8sControlPlane{}
			{
				m.Log("level", "debug", "message", fmt.Sprintf("Fetching G8sControlPlane %s", awsControlPlaneCR.Name))
				err := m.k8sClient.CtrlClient().Get(
					ctx,
					types.NamespacedName{Name: awsControlPlaneCR.GetName(), Namespace: namespace},
					g8sControlPlane,
				)
				if err != nil {
					return microerror.Maskf(aws.NotFoundError, "failed to fetch G8sControlplane: %v", err)
				}
			}
			numberOfAZs = g8sControlPlane.Spec.Replicas
			{
				// If the infrastructure reference is not set, we do it here
				if request.Operation == aws.CreateOperation && g8sControlPlane.Spec.InfrastructureRef.Name == "" {
					m.Log("level", "debug", "message", fmt.Sprintf("Updating infrastructure reference to  %s", awsControlPlaneCR.Name))
					infrastructureCRRef, err := reference.GetReference(infrastructurev1alpha2scheme.Scheme, awsControlPlaneCR)
					if infrastructureCRRef.Namespace == "" {
						infrastructureCRRef.Namespace = namespace
					}
					if err != nil {
						return microerror.Mask(err)
					}

					// We update the reference in the CR
					g8sControlPlane.Spec.InfrastructureRef = *infrastructureCRRef
					err = m.k8sClient.CtrlClient().Update(ctx, g8sControlPlane)
					if err != nil {
						return microerror.Mask(err)
					}
				}
			}
			return nil
		}
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		// Note that while we do log the error, we don't fail if the g8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		if aws.IsNotFound(err) {
			m.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlaneCR.Name, err))
		} else if err != nil {
			return nil, err
		}
	}
	if aws.IsHAVersion(releaseVersion) {
		// Trigger defaulting of the master instance type
		if awsControlPlaneCR.Spec.InstanceType == "" {
			m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s InstanceType is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			patch := mutator.PatchAdd("/spec/instanceType", aws.DefaultMasterInstanceType)
			result = append(result, patch)
		}
		// Trigger defaulting of the master availability zones
		if awsControlPlaneCR.Spec.AvailabilityZones == nil {
			m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			// We default the AZs
			defaultedAZs := m.getNavailabilityZones(numberOfAZs, m.validAvailabilityZones)
			patch := mutator.PatchAdd("/spec/availabilityZones", defaultedAZs)
			result = append(result, patch)
		}
	} else {
		var availabilityZone []string
		var instanceType string
		fetch := func() error {
			ctx := context.Background()

			// We try to fetch the AWSCluster CR.
			AWSCluster := &infrastructurev1alpha2.AWSCluster{}
			clusterID, err := clusterID(awsControlPlaneCR)
			if err != nil {
				return err
			}
			{
				m.Log("level", "debug", "message", fmt.Sprintf("Fetching AWSCluster %s", clusterID))
				err := m.k8sClient.CtrlClient().Get(ctx,
					types.NamespacedName{Name: clusterID,
						Namespace: namespace},
					AWSCluster)
				if err != nil {
					return microerror.Maskf(aws.NotFoundError, "failed to fetch AWSCluster: %v", err)
				}
			}
			availabilityZone = append(availabilityZone, AWSCluster.Spec.Provider.Master.AvailabilityZone)
			instanceType = AWSCluster.Spec.Provider.Master.InstanceType
			return nil
		}
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		if err != nil {
			m.Log("level", "debug", "message", fmt.Sprintf("No AWSCluster for AWSControlPlane %s could be found: %v", awsControlPlaneCR.Name, err))
		}
		// Trigger defaulting of the master instance type
		if awsControlPlaneCR.Spec.InstanceType == "" {
			m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s InstanceType is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			patch := mutator.PatchAdd("/spec/instanceType", instanceType)
			result = append(result, patch)
		}
		// Trigger defaulting of the master availability zone
		if awsControlPlaneCR.Spec.AvailabilityZones == nil {
			m.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			patch := mutator.PatchAdd("/spec/availabilityZones", availabilityZone)
			result = append(result, patch)
		}
	}

	return result, nil
}

func (m *Mutator) getNavailabilityZones(n int, azs []string) []string {
	randomAZs := azs
	// In case there are not enough distinct AZs, we repeat them
	for len(randomAZs) < n {
		randomAZs = append(randomAZs, azs...)
	}
	// We shuffle the AZs, pick the first n and sort them alphabetically
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(randomAZs), func(i, j int) { randomAZs[i], randomAZs[j] = randomAZs[j], randomAZs[i] })
	randomAZs = randomAZs[:n]
	sort.Strings(randomAZs)
	m.Log("level", "debug", "message", fmt.Sprintf("available AZ's: %v, selected AZ's: %v", azs, randomAZs))

	return randomAZs
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "awscontrolplane"
}

func releaseVersion(cr *infrastructurev1alpha2.AWSControlPlane) (*semver.Version, error) {
	version, ok := cr.Labels[label.ReleaseVersion]
	if !ok {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to get release version from AWSControlplane %s", cr.Name)
	}

	return semver.New(version)
}

func clusterID(cr *infrastructurev1alpha2.AWSControlPlane) (string, error) {
	clusterID, ok := cr.Labels[label.Cluster]
	if !ok {
		return "", microerror.Maskf(aws.ParsingFailedError, "unable to get cluster ID from AWSControlplane %s", cr.Name)
	}

	return clusterID, nil
}
