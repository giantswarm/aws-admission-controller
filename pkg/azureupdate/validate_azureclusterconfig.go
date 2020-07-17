package azureupdate

import (
	"fmt"

	"github.com/blang/semver"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/admission-controller/pkg/validator"
)

type AzureClusterConfigValidator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

type AzureClusterConfigValidatorConfig struct {
	Logger micrologger.Logger
}

func NewAzureClusterConfigValidator(config AzureClusterConfigValidatorConfig) (*AzureClusterConfigValidator, error) {
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

	validator := &AzureClusterConfigValidator{
		k8sClient: k8sClient,
		logger:    config.Logger,
	}

	return validator, nil
}

func (a *AzureClusterConfigValidator) Validate(request *v1beta1.AdmissionRequest) (bool, error) {
	AzureClusterConfigNewCR := &corev1alpha1.AzureClusterConfig{}
	AzureClusterConfigOldCR := &corev1alpha1.AzureClusterConfig{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, AzureClusterConfigNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse AzureClusterConfig CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, AzureClusterConfigOldCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse AzureClusterConfig CR: %v", err)
	}

	oldVersion, err := clusterConfigVersion(AzureClusterConfigOldCR)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse version from AzureClusterConfig (before edit)")
	}
	newVersion, err := clusterConfigVersion(AzureClusterConfigNewCR)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse version from AzureClusterConfig (after edit)")
	}

	if !oldVersion.Equals(*newVersion) {
		// The AzureClusterConfig CR doesn't have an indication of the fact that an update is in progress.
		// I need to use the corresponding AzureConfig CR for this check.
		acName := AzureClusterConfigOldCR.Spec.Guest.ID
		ac, err := a.k8sClient.G8sClient().ProviderV1alpha1().AzureConfigs("default").Get(acName, v1.GetOptions{})
		if err != nil {
			return false, microerror.Maskf(invalidOperationError, "Unable to find AzureConfig %s. Can't reliably tell if the cluster upgrade is safe or not. Error was %s", acName, err)
		}

		upgrading, status := clusterIsUpgrading(ac)
		if upgrading {
			return false, microerror.Maskf(invalidOperationError, "cluster has condition: %s", status)
		}

		return upgradeAllowed(a.k8sClient.G8sClient(), oldVersion, newVersion)
	}

	return true, nil
}

func (a *AzureClusterConfigValidator) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func clusterConfigVersion(cr *corev1alpha1.AzureClusterConfig) (*semver.Version, error) {
	version := cr.Spec.Guest.ReleaseVersion

	return semver.New(version)
}
