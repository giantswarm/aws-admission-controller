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
	"github.com/giantswarm/apiextensions/pkg/label"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	restclient "k8s.io/client-go/rest"
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

// var (
//  awsControlPlaneResource = metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "awscontrolplane"}
// )

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
	// Trigger defaulting of the master availability zones
	if awsControlPlaneCR.Spec.AvailabilityZones == nil && aws.IsHAVersion(releaseVersion) {
		a.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s AvailabilityZones is nil and will be defaulted", awsControlPlaneCR.ObjectMeta.Name))
		numberOfAZs := aws.DefaultMasterReplicas
		fetch := func() error {

			ctx := context.Background()

			// We try to fetch the G8sControlPlane CR.
			g8sControlPlane := &infrastructurev1alpha2.G8sControlPlane{}
			{
				a.Log("level", "debug", "message", fmt.Sprintf("Fetching G8sControlPlane %s", awsControlPlaneCR.Name))
				err := a.k8sClient.CtrlClient().Get(ctx,
					types.NamespacedName{Name: awsControlPlaneCR.GetName(),
						Namespace: awsControlPlaneCR.GetNamespace()},
					g8sControlPlane)
				if err != nil {
					return microerror.Maskf(aws.NotFoundError, "failed to fetch G8sControlplane: %v", err)
				}
			}
			numberOfAZs = g8sControlPlane.Spec.Replicas
			return nil
		}
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err := backoff.Retry(fetch, b)
		if err != nil {
			a.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found", awsControlPlaneCR.Name))
		}
		// We default the AZs
		patch := admission.PatchReplace("/spec/AvailabilityZones", a.getNavailabilityZones(numberOfAZs, a.validAvailabilityZones))
		result = append(result, patch)
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
	a.Log("level", "debug", "message", fmt.Sprintf("%d AZ's available, selected AZ's: %v", len(azs), randomAZs))

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
