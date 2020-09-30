package g8scontrolplane

import (
	"context"
	"fmt"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/config"
	"github.com/giantswarm/aws-admission-controller/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/pkg/key"
	"github.com/giantswarm/aws-admission-controller/pkg/validator"
)

type Validator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewValidator(config config.Config) (*Validator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(aws.InvalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(aws.InvalidConfigError, "%T.Logger must not be empty", config)
	}
	validator := &Validator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return validator, nil
}

func (v *Validator) Validate(request *v1beta1.AdmissionRequest) (bool, error) {
	var g8sControlPlane infrastructurev1alpha2.G8sControlPlane
	var allowed bool

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &g8sControlPlane); err != nil {
		return false, microerror.Maskf(aws.ParsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	allowed, err := v.ReplicaCount(g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	allowed, err = v.ReplicaAZMatch(g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}

	return allowed, nil
}

func (v *Validator) ReplicaAZMatch(g8sControlPlane infrastructurev1alpha2.G8sControlPlane) (bool, error) {
	var awsControlPlane infrastructurev1alpha2.AWSControlPlane
	var err error
	var fetch func() error

	// Fetch the AWSControlPlane.
	{
		v.Log("level", "debug", "message", fmt.Sprintf("Fetching AWSControlPlane %s", g8sControlPlane.Name))
		fetch = func() error {
			ctx := context.Background()

			err = v.k8sClient.CtrlClient().Get(
				ctx,
				types.NamespacedName{Name: g8sControlPlane.GetName(), Namespace: g8sControlPlane.GetNamespace()},
				&awsControlPlane,
			)
			if err != nil {
				return microerror.Maskf(aws.NotFoundError, "failed to fetch AWSControlplane: %v", err)
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

	if g8sControlPlane.Spec.Replicas != len(awsControlPlane.Spec.AvailabilityZones) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s with %v replicas does not match AWSControlPlane %s with %v availability zones %s",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			awsControlPlane.Spec.AvailabilityZones),
		)
		return false, microerror.Maskf(aws.NotAllowedError, fmt.Sprintf("G8sControlPlane %s with %v replicas does not match AWSControlPlane %s with %v availability zones %s",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			awsControlPlane.Spec.AvailabilityZones),
		)
	}

	return true, nil
}
func (v *Validator) ReplicaCount(g8sControlPlane infrastructurev1alpha2.G8sControlPlane) (bool, error) {
	if !aws.IsValidMasterReplicas(g8sControlPlane.Spec.Replicas) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s has an invalid count of %v replicas. Valid replica counts are: %v",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			aws.ValidMasterReplicas()),
		)
		return false, microerror.Maskf(aws.NotAllowedError, fmt.Sprintf("G8sControlPlane %s has an invalid count of %v replicas. Valid replica counts are: %v",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			aws.ValidMasterReplicas()),
		)
	}

	return true, nil
}

func (m *Validator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}
