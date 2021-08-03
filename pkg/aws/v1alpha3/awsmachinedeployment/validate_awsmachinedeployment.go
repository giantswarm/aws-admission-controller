package awsmachinedeployment

import (
	"context"
	"fmt"
	"strings"
	"time"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/aws-admission-controller/v3/config"
	aws "github.com/giantswarm/aws-admission-controller/v3/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/validator"
)

type Validator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	validInstanceTypes []string
}

func NewValidator(config config.Config) (*Validator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var instanceTypes []string = strings.Split(config.WorkerInstanceTypes, ",")

	validator := &Validator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		validInstanceTypes: instanceTypes,
	}

	return validator, nil
}

func (v *Validator) Validate(request *admissionv1.AdmissionRequest) (bool, error) {
	if request.Operation == admissionv1.Update {
		return v.ValidateUpdate(request)
	}
	if request.Operation == admissionv1.Create {
		return v.ValidateCreate(request)
	}
	return true, nil
}

func (v *Validator) ValidateUpdate(request *admissionv1.AdmissionRequest) (bool, error) {
	var awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsMachineDeployment); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awsmachinedeployment: %v", err)
	}

	err = v.InstanceTypeValid(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentLabelMatch(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentAnnotationMaxBatchSizeIsValid(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentAnnotationPauseTimeIsValid(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentScaling(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) ValidateCreate(request *admissionv1.AdmissionRequest) (bool, error) {
	var err error

	var awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsMachineDeployment); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awsmachinedeployment: %v", err)
	}

	err = aws.ValidateOrganizationLabelContainsExistingOrganization(context.Background(), v.k8sClient.CtrlClient(), &awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.InstanceTypeValid(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.ValidateCluster(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentLabelMatch(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentAnnotationMaxBatchSizeIsValid(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentAnnotationPauseTimeIsValid(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.MachineDeploymentScaling(awsMachineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) InstanceTypeValid(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) error {
	if !contains(v.validInstanceTypes, awsMachineDeployment.Spec.Provider.Worker.InstanceType) {
		return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSMachineDeployment %s worker instance type %v is invalid. Valid instance types are: %v",
			key.ControlPlane(&awsMachineDeployment),
			awsMachineDeployment.Spec.Provider.Worker.InstanceType,
			v.validInstanceTypes),
		)
	}

	return nil
}

func (v *Validator) MachineDeploymentLabelMatch(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) error {
	var machineDeployment v1alpha3.MachineDeployment
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
				return microerror.Maskf(notFoundError, "failed to fetch MachineDeployment: %v", err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 10*time.Millisecond)
		err = backoff.Retry(fetch, b)
		// Note that while we do log the error, we don't fail if the MachineDeployment doesn't exist yet. That is okay because the order of CR creation can vary.
		if IsNotFound(err) {
			v.Log("level", "debug", "message", fmt.Sprintf("No MachineDeployment %s could be found: %v", awsMachineDeployment.GetName(), err))
			return nil
		} else if err != nil {
			return microerror.Mask(err)
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
		return microerror.Maskf(notAllowedError, fmt.Sprintf("MachineDeployment %s=%s label does not match with AWSMachineDeployment %s=%s label for cluster %s",
			label.MachineDeployment,
			key.MachineDeployment(&machineDeployment),
			label.MachineDeployment,
			key.MachineDeployment(&awsMachineDeployment),
			key.Cluster(&awsMachineDeployment)),
		)
	}

	return nil
}

func (v *Validator) MachineDeploymentAnnotationMaxBatchSizeIsValid(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) error {
	if maxBatchSize, ok := awsMachineDeployment.GetAnnotations()[aws.AnnotationUpdateMaxBatchSize]; ok {
		if !aws.MaxBatchSizeIsValid(maxBatchSize) {
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSMachineDeployment annotation '%s' value '%s' is not valid. Allowed value is either integer bigger than zero or decimal number between 0 and 1.0 defining percentage of nodes",
				aws.AnnotationUpdateMaxBatchSize,
				maxBatchSize),
			)
		}
	}
	return nil
}

func (v *Validator) MachineDeploymentAnnotationPauseTimeIsValid(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) error {
	if maxBatchSize, ok := awsMachineDeployment.GetAnnotations()[aws.AnnotationUpdatePauseTime]; ok {
		if !aws.PauseTimeIsValid(maxBatchSize) {
			v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSMachineDeployment annotation '%s' value '%s' is not valid. Value must be in ISO 8601 duration format and cannot be bigger than 1 hour.",
				aws.AnnotationUpdatePauseTime,
				maxBatchSize),
			)
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSMachineDeployment annotation '%s' value '%s' is not valid. Value must be in ISO 8601 duration format and cannot be bigger than 1 hour.",
				aws.AnnotationUpdatePauseTime,
				maxBatchSize),
			)
		}
	}
	return nil
}

func (v *Validator) ValidateCluster(awsMachineDeployment infrastructurev1alpha3.AWSMachineDeployment) error {
	var err error

	// Retrieve the `Cluster` CR related to this object.
	cluster, err := aws.FetchCluster(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, &awsMachineDeployment)
	if err != nil {
		return microerror.Mask(err)
	}
	// make sure the cluster is not deleted
	if cluster.DeletionTimestamp != nil {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSMachineDeployment could not be created because Cluster '%s' is in deleting state.",
			cluster.Name),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSMachineDeployment could not be created because Cluster '%s' is in deleting state.",
			cluster.Name),
		)
	}
	return nil
}

func (v *Validator) MachineDeploymentScaling(md infrastructurev1alpha3.AWSMachineDeployment) error {
	min := md.Spec.NodePool.Scaling.Min
	max := md.Spec.NodePool.Scaling.Max

	if max == 0 {
		return microerror.Maskf(notAllowedError, "AWSMachineDeployment.Spec.Scaling.Max must not be 0.")
	}

	if min > max {
		return microerror.Maskf(notAllowedError, "AWSMachineDeployment.Spec.Scaling.Min must not be greater that AWSMachineDeployment.Spec.Scaling.Max.")
	}

	return nil
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "awsmachinedeployment"
}
