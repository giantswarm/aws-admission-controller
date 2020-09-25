package awscontrolplane

import (
	"context"
	"fmt"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/pkg/key"
	"github.com/giantswarm/aws-admission-controller/pkg/label"
	"github.com/giantswarm/aws-admission-controller/pkg/validator"
)

type Validator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewValidator() (*Validator, error) {
	var err error
	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

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
			Logger: newLogger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	validator := &Validator{
		k8sClient: k8sClient,
		logger:    newLogger,
	}

	return validator, nil
}

func (v *Validator) Validate(request *v1beta1.AdmissionRequest) (bool, error) {
	var awsControlPlane infrastructurev1alpha2.AWSControlPlane
	var allowed bool

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsControlPlane); err != nil {
		return false, microerror.Maskf(aws.ParsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	allowed, err := v.ControlPlaneLabelMatch(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}

	return allowed, nil
}

func (v *Validator) ControlPlaneLabelMatch(awsControlPlane infrastructurev1alpha2.AWSControlPlane) (bool, error) {
	var g8sControlPlane infrastructurev1alpha2.G8sControlPlane
	var err error
	var fetch func() error

	// Fetch the G8sControlPlane.
	{
		v.Log("level", "debug", "message", fmt.Sprintf("Fetching G8sControlPlane %s", awsControlPlane.Name))
		fetch = func() error {
			ctx := context.Background()

			err = v.k8sClient.CtrlClient().Get(
				ctx,
				types.NamespacedName{Name: awsControlPlane.GetName(), Namespace: awsControlPlane.GetNamespace()},
				&g8sControlPlane,
			)
			if err != nil {
				return microerror.Maskf(aws.NotFoundError, "failed to fetch G8sControlplane: %v", err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	if key.ControlPlane(&g8sControlPlane) != key.ControlPlane(&awsControlPlane) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s=%s label does not match with AWSControlPlane %s=%s label for cluster %s",
			label.ControlPlane,
			key.ControlPlane(&g8sControlPlane),
			label.ControlPlane,
			key.ControlPlane(&awsControlPlane),
			key.Cluster(&g8sControlPlane)),
		)
		return false, nil
	}

	return true, nil
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}
