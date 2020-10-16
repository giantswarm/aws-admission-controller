package awscontrolplane

import (
	"context"
	"fmt"
	"strings"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
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

	validAvailabilityZones []string
	validInstanceTypes     []string
}

func NewValidator(config config.Config) (*Validator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var availabilityZones []string = strings.Split(config.AvailabilityZones, ",")
	var instanceTypes []string = strings.Split(config.MasterInstanceTypes, ",")

	validator := &Validator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		validAvailabilityZones: availabilityZones,
		validInstanceTypes:     instanceTypes,
	}

	return validator, nil
}

func (v *Validator) Validate(request *admissionv1.AdmissionRequest) (bool, error) {
	var awsControlPlane infrastructurev1alpha2.AWSControlPlane

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsControlPlane); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	azAllowed, err := v.AZValid(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	azCountAllowed, err := v.AZCount(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	instanceTypeAllowed, err := v.InstanceTypeValid(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	azOrderKept, err := v.AZOrder(request)
	if err != nil {
		return false, microerror.Mask(err)

	}
	azUnique, err := v.AZUnique(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	controlPlaneLabelMatches, err := v.ControlPlaneLabelMatch(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}
	azReplicaMatches, err := v.AZReplicaMatch(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)

	}

	return controlPlaneLabelMatches && azAllowed && azCountAllowed && azUnique && azOrderKept && azReplicaMatches && instanceTypeAllowed, nil
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
				return microerror.Maskf(notFoundError, "failed to fetch g8sControlplane: %v", err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		if IsNotFound(err) {
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
		return false, microerror.Maskf(notAllowedError, fmt.Sprintf("G8sControlPlane %s with %v replicas does not match AWSControlPlane %s with %v availability zones %s",
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
		return false, microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s has an invalid count of %v availability zones. Valid AZ counts are: %v",
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			aws.ValidMasterReplicas()),
		)
	}

	return true, nil
}
func (v *Validator) AZOrder(request *v1beta1.AdmissionRequest) (bool, error) {
	// Order can only change on update
	if request.Operation != aws.UpdateOperation {
		return true, nil
	}
	var awsControlPlane infrastructurev1alpha2.AWSControlPlane
	var awsControlPlaneOld infrastructurev1alpha2.AWSControlPlane
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsControlPlane); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, &awsControlPlaneOld); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse old awscontrol plane: %v", err)
	}
	if orderChanged(awsControlPlaneOld.Spec.AvailabilityZones, awsControlPlane.Spec.AvailabilityZones) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s order of AZs has changed from %v to %v.",
			key.ControlPlane(&awsControlPlane),
			awsControlPlaneOld.Spec.AvailabilityZones,
			awsControlPlane.Spec.AvailabilityZones),
		)
		return false, microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s order of AZs has changed from %v to %v.",
			key.ControlPlane(&awsControlPlane),
			awsControlPlaneOld.Spec.AvailabilityZones,
			awsControlPlane.Spec.AvailabilityZones),
		)
	}
	return true, nil
}
func (v *Validator) AZUnique(awsControlPlane infrastructurev1alpha2.AWSControlPlane) (bool, error) {
	// We always want to select as many distinct AZs as possible
	distinctAZs := countUniqueValues(awsControlPlane.Spec.AvailabilityZones)
	if distinctAZs == len(v.validAvailabilityZones) || distinctAZs == len(awsControlPlane.Spec.AvailabilityZones) {
		return true, nil
	}
	v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s availability zones %v do not contain maximum amount of distinct AZs. Valid AZs are: %v",
		key.ControlPlane(&awsControlPlane),
		awsControlPlane.Spec.AvailabilityZones,
		v.validAvailabilityZones),
	)
	return false, microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s availability zones %v do not contain maximum amount of distinct AZs. Valid AZs are: %v",
		key.ControlPlane(&awsControlPlane),
		awsControlPlane.Spec.AvailabilityZones,
		v.validAvailabilityZones),
	)
}

func (v *Validator) AZValid(awsControlPlane infrastructurev1alpha2.AWSControlPlane) (bool, error) {
	if !v.isValidMasterAvailabilityZones(awsControlPlane.Spec.AvailabilityZones) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s availability zones %v are invalid. Valid AZs are: %v",
			key.ControlPlane(&awsControlPlane),
			awsControlPlane.Spec.AvailabilityZones,
			v.validAvailabilityZones),
		)
		return false, microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s availability zones %v are invalid. Valid AZs are: %v",
			key.ControlPlane(&awsControlPlane),
			awsControlPlane.Spec.AvailabilityZones,
			v.validAvailabilityZones),
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
				return microerror.Maskf(notFoundError, "failed to fetch G8sControlplane: %v", err)
			}
			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		if IsNotFound(err) {
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
		return false, microerror.Maskf(notAllowedError, fmt.Sprintf("G8sControlPlane %s=%s label does not match with AWSControlPlane %s=%s label for cluster %s",
			label.ControlPlane,
			key.ControlPlane(&g8sControlPlane),
			label.ControlPlane,
			key.ControlPlane(&awsControlPlane),
			key.Cluster(&g8sControlPlane)),
		)
	}

	return true, nil
}
func (v *Validator) InstanceTypeValid(awsControlPlane infrastructurev1alpha2.AWSControlPlane) (bool, error) {
	if !contains(v.validInstanceTypes, awsControlPlane.Spec.InstanceType) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s master instance type %v is invalid. Valid instance types are: %v",
			key.ControlPlane(&awsControlPlane),
			awsControlPlane.Spec.InstanceType,
			v.validInstanceTypes),
		)
		return false, microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s master instance type %v is invalid. Valid instance types are: %v",
			key.ControlPlane(&awsControlPlane),
			awsControlPlane.Spec.InstanceType,
			v.validInstanceTypes),
		)
	}

	return true, nil
}

func (v *Validator) isValidMasterAvailabilityZones(availabilityZones []string) bool {
	for _, az := range availabilityZones {
		if !contains(v.validAvailabilityZones, az) {
			return false
		}
	}
	return true
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func countUniqueValues(s []string) int {
	counter := make(map[string]int)
	for _, a := range s {
		counter[a]++
	}
	return len(counter)
}
func orderChanged(old []string, new []string) bool {
	if len(old) > len(new) {
		temp := old
		old = new
		new = temp
	}
	for i, o := range old {
		for _, n := range new {
			if o == n && o != new[i] {
				return true
			}
		}
	}
	return false
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "awscontrolplane"
}
