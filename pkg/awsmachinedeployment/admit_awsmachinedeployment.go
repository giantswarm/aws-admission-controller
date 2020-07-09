// Package awsmachinedeployment intercepts write activity to AWSMachineDeployment objects.
package awsmachinedeployment

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/admission-controller/pkg/admission"
)

var (
	// If not specified otherwise, node pools should have 100% on-demand instances.
	defaultOnDemandPercentageAboveBaseCapacity int = 100
)

// Admitter defines our admitter object.
type Admitter struct {
	k8sClient k8sclient.Interface
}

// AdmitterConfig configures our Admitter.
type AdmitterConfig struct {
}

// NewAdmitter returns a new admitter.
func NewAdmitter(cfg *AdmitterConfig) (*Admitter, error) {
	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load key kubeconfig: %v", err)
		}
		newLogger, err := micrologger.New(micrologger.Config{})
		if err != nil {
			return nil, err
		}
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				apiv1alpha2.AddToScheme,
				infrastructurev1alpha2.AddToScheme,
				releasev1alpha1.AddToScheme,
			},
			Logger: newLogger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	admitter := &Admitter{
		k8sClient: k8sClient,
	}

	return admitter, nil
}

// Admit is the function executed for every matching webhook request.
func (admitter *Admitter) Admit(request *v1beta1.AdmissionRequest) ([]admission.PatchOperation, error) {
	// Parse incoming objects
	awsMachineDeploymentNewCR := &infrastructurev1alpha2.AWSMachineDeployment{}
	awsMachineDeploymentOldCR := &infrastructurev1alpha2.AWSMachineDeployment{}
	if _, _, err := admission.Deserializer.Decode(request.Object.Raw, nil, awsMachineDeploymentNewCR); err != nil {
		log.Errorf("unable to parse AWSMachineDeployment: %v", err)
		return nil, admission.InternalError
	}
	if _, _, err := admission.Deserializer.Decode(request.OldObject.Raw, nil, awsMachineDeploymentOldCR); err != nil {
		log.Errorf("unable to parse AWSMachineDeployment: %v", err)
		return nil, admission.InternalError
	}

	var result []admission.PatchOperation

	// Default the OnDemandPercentageAboveBaseCapacity.
	// Note: This will only work if the incoming CR has the .spec.provider.instanceDistribution
	// attribute defined. Otherwise the request to create/modify the CR will fail.
	if awsMachineDeploymentNewCR.Spec.Provider.InstanceDistribution.OnDemandPercentageAboveBaseCapacity == nil {
		log.Infof("AWSMachineDeployment %s onDemandBaseCapacity is nil and will be set to default 100", awsMachineDeploymentNewCR.ObjectMeta.Name)
		patch := admission.PatchReplace("/spec/provider/instanceDistribution/onDemandPercentageAboveBaseCapacity", &defaultOnDemandPercentageAboveBaseCapacity)
		result = append(result, patch)
	}

	return result, nil
}
