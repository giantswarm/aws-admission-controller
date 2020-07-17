package azureupdate

import (
	"fmt"

	"github.com/blang/semver"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/admission-controller/pkg/validator"
)

type AzureConfigAdmitter struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

type AzureConfigAdmitterConfig struct {
	Logger micrologger.Logger
}

const (
	conditionCreating = "Creating"
	conditionUpdating = "Updating"
	versionLabel      = "release.giantswarm.io/version"
)

func NewAzureConfigAdmitter(config AzureConfigAdmitterConfig) (*AzureConfigAdmitter, error) {
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

	admitter := &AzureConfigAdmitter{
		k8sClient: k8sClient,
		logger:    config.Logger,
	}

	return admitter, nil
}

func (a *AzureConfigAdmitter) Validate(request *v1beta1.AdmissionRequest) (bool, error) {
	azureConfigNewCR := &v1alpha1.AzureConfig{}
	azureConfigOldCR := &v1alpha1.AzureConfig{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, azureConfigNewCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse azureConfig CR: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, azureConfigOldCR); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse azureConfig CR: %v", err)
	}

	oldVersion, err := clusterVersion(azureConfigOldCR)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse version from AzureConfig (before edit)")
	}
	newVersion, err := clusterVersion(azureConfigNewCR)
	if err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse version from AzureConfig (after edit)")
	}

	if !oldVersion.Equals(*newVersion) {
		// If tenant cluster is already upgrading, we can't change the version any more.
		upgrading, status := clusterIsUpgrading(azureConfigOldCR)
		if upgrading {
			return false, microerror.Maskf(invalidOperationError, "cluster has condition: %s", status)
		}

		return upgradeAllowed(a.k8sClient.G8sClient(), oldVersion, newVersion)
	}

	return true, nil
}

func (a *AzureConfigAdmitter) Log(keyVals ...interface{}) {
	a.logger.Log(keyVals...)
}

func clusterIsUpgrading(cr *v1alpha1.AzureConfig) (bool, string) {
	for _, cond := range cr.Status.Cluster.Conditions {
		if cond.Type == conditionUpdating {
			return true, conditionUpdating
		}
		if cond.Type == conditionCreating {
			return true, conditionCreating
		}
	}

	return false, ""
}

func clusterVersion(cr *v1alpha1.AzureConfig) (*semver.Version, error) {
	version, ok := cr.Labels[versionLabel]
	if !ok {
		return nil, microerror.Maskf(parsingFailedError, "unable to get cluster version from AzureConfig %s", cr.Name)
	}

	return semver.New(version)
}
