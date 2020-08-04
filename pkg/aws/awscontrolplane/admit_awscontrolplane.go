package awscontrolplane

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	infrastructurev1alpha2scheme "github.com/giantswarm/apiextensions/pkg/clientset/versioned/scheme"
	"github.com/giantswarm/apiextensions/pkg/label"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/reference"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/admission-controller/pkg/admission"
	"github.com/giantswarm/admission-controller/pkg/aws"
)

type Config struct {
	ValidAvailabilityZones string
	Logger                 micrologger.Logger
}

type Admitter struct {
	k8sClient              k8sclient.Interface
	validAvailabilityZones []string
	logger                 micrologger.Logger
}

func NewAdmitter(config Config) (*Admitter, error) {
	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load key kubeconfig: %v", err)
		}
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				apiv1alpha2.AddToScheme,
				infrastructurev1alpha2.AddToScheme,
				releasev1alpha1.AddToScheme,
			},
			Logger: config.Logger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var availabilityZones []string = strings.Split(config.ValidAvailabilityZones, ",")
	admitter := &Admitter{
		k8sClient:              k8sClient,
		validAvailabilityZones: availabilityZones,
		logger:                 config.Logger,
	}

	return admitter, nil
}

func (a *Admitter) Admit(request *v1beta1.AdmissionRequest) ([]admission.PatchOperation, error) {
	awsControlPlaneCR := &infrastructurev1alpha2.AWSControlPlane{}
	if _, _, err := admission.Deserializer.Decode(request.Object.Raw, nil, awsControlPlaneCR); err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	releaseVersion, err := releaseVersion(awsControlPlaneCR)
	if err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse release version from AWSControlPlane")
	}

	var result []admission.PatchOperation
	var numberOfAZs int

	// We only need to manipulate if attributes are not set or it's a create operation
	if awsControlPlaneCR.Spec.AvailabilityZones != nil && awsControlPlaneCR.Spec.InstanceType != "" && request.Operation != "CREATE" {
		return result, nil
	}
	// We need to fetch the G8sControlPlane in case AZs need to be defaulted or the awscontrolplane is just created
	if (aws.IsHAVersion(releaseVersion) && awsControlPlaneCR.Spec.AvailabilityZones == nil) || request.Operation == "CREATE" {
		numberOfAZs = aws.DefaultMasterReplicas
		fetch := func() error {
			ctx := context.Background()

			// We try to fetch the G8sControlPlane CR.
			g8sControlPlane := &infrastructurev1alpha2.G8sControlPlane{}
			{
				a.Log("level", "debug", "message", fmt.Sprintf("Fetching G8sControlPlane %s", awsControlPlaneCR.Name))
				err := a.k8sClient.CtrlClient().Get(
					ctx,
					types.NamespacedName{Name: awsControlPlaneCR.GetName(), Namespace: awsControlPlaneCR.GetNamespace()},
					g8sControlPlane,
				)
				if err != nil {
					return microerror.Maskf(aws.NotFoundError, "failed to fetch G8sControlplane: %v", err)
				}
			}
			numberOfAZs = g8sControlPlane.Spec.Replicas
			{
				// If the infrastructure reference is not set, we do it here
				if request.Operation == "CREATE" && g8sControlPlane.Spec.InfrastructureRef.Name == "" {
					a.Log("level", "debug", "message", fmt.Sprintf("Updating infrastructure reference to  %s", awsControlPlaneCR.Name))
					infrastructureCRRef, err := reference.GetReference(infrastructurev1alpha2scheme.Scheme, awsControlPlaneCR)
					if err != nil {
						return microerror.Maskf(aws.ExecutionFailedError, "failed to create reference to AWSControlplane: %v", err)
					}

					// We update the reference in the CR
					g8sControlPlane.Spec.InfrastructureRef = *infrastructureCRRef
					err = a.k8sClient.CtrlClient().Update(ctx, g8sControlPlane)
					if err != nil {
						return microerror.Maskf(aws.ExecutionFailedError, "failed to update G8sControlplane: %v", err)
					}
				}
			}
			return nil
		}
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		// Note that while we do log the error, we don't fail if the g8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		if aws.IsNotFound(err) {
			a.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlaneCR.Name, err))
		} else if err != nil {
			return nil, err
		}
	}
	if aws.IsHAVersion(releaseVersion) {
		// Trigger defaulting of the master instance type
		if awsControlPlaneCR.Spec.InstanceType == "" {
			a.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s InstanceType is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			patch := admission.PatchAdd("/spec/instanceType", aws.DefaultMasterInstanceType)
			result = append(result, patch)
		}
		// Trigger defaulting of the master availability zones
		if awsControlPlaneCR.Spec.AvailabilityZones == nil {
			a.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			// We default the AZs
			defaultedAZs := a.getNavailabilityZones(numberOfAZs, a.validAvailabilityZones)
			patch := admission.PatchAdd("/spec/availabilityZones", defaultedAZs)
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
				a.Log("level", "debug", "message", fmt.Sprintf("Fetching AWSCluster %s", clusterID))
				err := a.k8sClient.CtrlClient().Get(ctx,
					types.NamespacedName{Name: clusterID,
						Namespace: awsControlPlaneCR.GetNamespace()},
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
			a.Log("level", "debug", "message", fmt.Sprintf("No AWSCluster for AWSControlPlane %s could be found: %v", awsControlPlaneCR.Name, err))
		}
		// Trigger defaulting of the master instance type
		if awsControlPlaneCR.Spec.InstanceType == "" {
			a.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s InstanceType is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			patch := admission.PatchAdd("/spec/instanceType", instanceType)
			result = append(result, patch)
		}
		// Trigger defaulting of the master availability zone
		if awsControlPlaneCR.Spec.AvailabilityZones == nil {
			a.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
			patch := admission.PatchAdd("/spec/availabilityZones", availabilityZone)
			result = append(result, patch)
		}
	}

	return result, nil
}

func (a *Admitter) getNavailabilityZones(n int, azs []string) []string {
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
	a.Log("level", "debug", "message", fmt.Sprintf("available AZ's: %v, selected AZ's: %v", azs, randomAZs))

	return randomAZs
}

func (a *Admitter) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
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
