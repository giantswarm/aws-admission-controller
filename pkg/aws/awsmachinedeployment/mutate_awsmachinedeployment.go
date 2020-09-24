// Package awsmachinedeployment intercepts write activity to AWSMachineDeployment objects.
package awsmachinedeployment

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
)

var (
	// If not specified otherwise, node pools should have 100% on-demand instances.
	defaultOnDemandPercentageAboveBaseCapacity int = 100
)

// Mutator for AWSMachineDeployment object.
type Mutator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

// Config configures AWSMachineDeployment Admitter.
type Config struct {
	Logger micrologger.Logger
}

// NewMutator returns a new mutator.
func NewMutator(config Config) (*Mutator, error) {
	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return nil, microerror.Mask(err)
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

	mutator := &Mutator{
		k8sClient: k8sClient,
		logger:    config.Logger,
	}

	return mutator, nil
}

// Mutate is the function executed for every matching webhook request.
func (m *Mutator) Mutate(request *v1beta1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	// Parse incoming objects
	awsMachineDeploymentNewCR := &infrastructurev1alpha2.AWSMachineDeployment{}
	awsMachineDeploymentOldCR := &infrastructurev1alpha2.AWSMachineDeployment{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsMachineDeploymentNewCR); err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse AWSMachineDeployment: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, awsMachineDeploymentOldCR); err != nil {
		return nil, microerror.Maskf(aws.ParsingFailedError, "unable to parse AWSMachineDeployment: %v", err)
	}

	var result []mutator.PatchOperation

	// Default the OnDemandPercentageAboveBaseCapacity.
	// Note: This will only work if the incoming CR has the .spec.provider.instanceDistribution
	// attribute defined. Otherwise the request to create/modify the CR will fail.
	if awsMachineDeploymentNewCR.Spec.Provider.InstanceDistribution.OnDemandPercentageAboveBaseCapacity == nil {
		m.Log("level", "debug", "message", fmt.Sprintf("AWSMachineDeployment %s OnDemandPercentageAboveBaseCapacity is nil and will be set to default 100", awsMachineDeploymentNewCR.ObjectMeta.Name))
		patch := mutator.PatchReplace("/spec/provider/instanceDistribution/onDemandPercentageAboveBaseCapacity", &defaultOnDemandPercentageAboveBaseCapacity)
		result = append(result, patch)
	}

	return result, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}