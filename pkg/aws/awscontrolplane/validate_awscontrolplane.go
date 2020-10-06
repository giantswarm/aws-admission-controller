package awscontrolplane

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
	"github.com/giantswarm/aws-admission-controller/pkg/label"
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
	var awsControlPlane infrastructurev1alpha2.AWSControlPlane

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsControlPlane); err != nil {
		return false, microerror.Maskf(aws.ParsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	controlPlaneLabelMatches, err := v.ControlPlaneLabelMatch(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	azCountAllowed, err := v.AZCount(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	azReplicaMatches, err := v.AZReplicaMatch(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}

	return controlPlaneLabelMatches && azCountAllowed && azReplicaMatches, nil
}

func (v *Validator) AZReplicaMatch(awsControlPlane infrastructurev1alpha2.AWSControlPlane) (bool, error) {
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
				return microerror.Maskf(aws.NotFoundError, "failed to fetch g8sControlplane: %v", err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		if aws.IsNotFound(err) {
			v.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlane.GetName(), err))
			return true, nil
		} else if err != nil {
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
func (v *Validator) AZCount(awsControlPlane infrastructurev1alpha2.AWSControlPlane) (bool, error) {
	if !aws.IsValidMasterReplicas(len(awsControlPlane.Spec.AvailabilityZones)) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s has an invalid count of %v availability zones. Valid AZ counts are: %v",
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			aws.ValidMasterReplicas()),
		)
		return false, microerror.Maskf(aws.NotAllowedError, fmt.Sprintf("AWSControlPlane %s has an invalid count of %v availability zones. Valid AZ counts are: %v",
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			aws.ValidMasterReplicas()),
		)
	}

	return true, nil
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
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		if aws.IsNotFound(err) {
			v.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlane.GetName(), err))
			return true, nil
		} else if err != nil {
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
		return false, microerror.Maskf(aws.NotAllowedError, fmt.Sprintf("G8sControlPlane %s=%s label does not match with AWSControlPlane %s=%s label for cluster %s",
			label.ControlPlane,
			key.ControlPlane(&g8sControlPlane),
			label.ControlPlane,
			key.ControlPlane(&awsControlPlane),
			key.Cluster(&g8sControlPlane)),
		)
	}

	return true, nil
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "awscontrolplane"
}
