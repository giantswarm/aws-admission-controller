package awsmachinedeployment

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
	"sigs.k8s.io/cluster-api/api/v1alpha2"

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
	var awsMachineDeployment infrastructurev1alpha2.AWSMachineDeployment
	var allowed bool

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsMachineDeployment); err != nil {
		return false, microerror.Maskf(aws.ParsingFailedError, "unable to parse awsmachinedeployment: %v", err)
	}
	allowed, err := v.MachineDeploymentLabelMatch(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)

	}

	return allowed, nil
}

func (v *Validator) MachineDeploymentLabelMatch(awsMachineDeployment infrastructurev1alpha2.AWSMachineDeployment) (bool, error) {
	var machineDeployment v1alpha2.MachineDeployment
	var err error
	var fetch func() error

	// Fetch the MachineDeployment.
	{
		v.Log("level", "debug", "message", fmt.Sprintf("Fetching MachineDeployment %s", awsMachineDeployment.Name))
		fetch = func() error {
			ctx := context.Background()

			err = v.k8sClient.CtrlClient().Get(
				ctx,
				types.NamespacedName{Name: awsMachineDeployment.GetName(), Namespace: awsMachineDeployment.GetNamespace()},
				&machineDeployment,
			)
			if err != nil {
				return microerror.Maskf(aws.NotFoundError, "failed to fetch MachineDeployment: %v", err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		// Note that while we do log the error, we don't fail if the MachineDeployment doesn't exist yet. That is okay because the order of CR creation can vary.
		if aws.IsNotFound(err) {
			v.Log("level", "debug", "message", fmt.Sprintf("No MachineDeployment %s could be found: %v", awsMachineDeployment.GetName(), err))
			return true, nil
		} else if err != nil {
			return false, microerror.Mask(err)
		}
	}

	if key.MachineDeployment(&machineDeployment) != key.MachineDeployment(&awsMachineDeployment) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("MachineDeployment %s=%s label does not match with AWSMachineDeployment %s=%s label for cluster %s",
			label.MachineDeployment,
			key.MachineDeployment(&machineDeployment),
			label.MachineDeployment,
			key.MachineDeployment(&awsMachineDeployment),
			key.Cluster(&awsMachineDeployment)),
		)
		return false, microerror.Maskf(aws.NotAllowedError, fmt.Sprintf("MachineDeployment %s=%s label does not match with AWSMachineDeployment %s=%s label for cluster %s",
			label.MachineDeployment,
			key.MachineDeployment(&machineDeployment),
			label.MachineDeployment,
			key.MachineDeployment(&awsMachineDeployment),
			key.Cluster(&awsMachineDeployment)),
		)
	}

	return true, nil
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "awsmachinedeployment"
}
