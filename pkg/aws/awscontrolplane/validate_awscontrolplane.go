package awscontrolplane

import (
	"context"
	"fmt"
	"strings"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/validator"
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
	var g8sControlPlane *infrastructurev1alpha2.G8sControlPlane
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsControlPlane); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscontrol plane: %v", err)
	}
	err = v.AZCount(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = v.AZValid(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = aws.ValidateOrganizationLabelContainsExistingOrganization(context.Background(), v.k8sClient.CtrlClient(), &awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}

	// The order can only change on update
	if request.Operation == admissionv1.Update {
		var awsControlPlaneOld infrastructurev1alpha2.AWSControlPlane
		if _, _, err := validator.Deserializer.Decode(request.OldObject.Raw, nil, &awsControlPlaneOld); err != nil {
			return false, microerror.Maskf(parsingFailedError, "unable to parse old awscontrol plane: %v", err)
		}
		err = v.AZOrder(awsControlPlane, awsControlPlaneOld)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}
	err = v.AZUnique(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = v.InstanceTypeValid(awsControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}
	// We try to fetch the G8sControlPlane belonging to the AWSControlPlane here.
	g8sControlPlane, err = aws.FetchG8sControlPlane(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, &awsControlPlane)
	if aws.IsNotFound(err) {
		// Note that while we do log the error, we don't fail if the G8sControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
		v.Log("level", "debug", "message", fmt.Sprintf("No G8sControlPlane %s could be found: %v", awsControlPlane.GetName(), err))
	} else if err != nil {
		return false, microerror.Mask(err)
	} else {
		// We only validate the matching of labels if we succeed in fetching the g8scontrolplane
		err = v.ControlPlaneLabelMatch(awsControlPlane, *g8sControlPlane)
		if err != nil {
			return false, microerror.Mask(err)
		}
		// when updating from single to HA validation of AZ replicas has to be ignored
		if request.Operation == admissionv1.Create {
			err = v.AZReplicaMatch(awsControlPlane, *g8sControlPlane)
			if err != nil {
				return false, microerror.Mask(err)
			}
		}
	}
	return true, nil
}

func (v *Validator) AZReplicaMatch(awsControlPlane infrastructurev1alpha2.AWSControlPlane, g8sControlPlane infrastructurev1alpha2.G8sControlPlane) error {
	if g8sControlPlane.Spec.Replicas != len(awsControlPlane.Spec.AvailabilityZones) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s with %v replicas does not match AWSControlPlane %s with %v availability zones %s",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			awsControlPlane.Spec.AvailabilityZones),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("G8sControlPlane %s with %v replicas does not match AWSControlPlane %s with %v availability zones %s",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			awsControlPlane.Spec.AvailabilityZones),
		)
	}

	return nil
}
func (v *Validator) AZCount(awsControlPlane infrastructurev1alpha2.AWSControlPlane) error {
	if !aws.IsValidMasterReplicas(len(awsControlPlane.Spec.AvailabilityZones)) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s has an invalid count of %v availability zones. Valid AZ counts are: %v",
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			aws.ValidMasterReplicas()),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s has an invalid count of %v availability zones. Valid AZ counts are: %v",
			key.ControlPlane(&awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			aws.ValidMasterReplicas()),
		)
	}

	return nil
}
func (v *Validator) AZOrder(awsControlPlane infrastructurev1alpha2.AWSControlPlane, awsControlPlaneOld infrastructurev1alpha2.AWSControlPlane) error {
	if orderChanged(awsControlPlaneOld.Spec.AvailabilityZones, awsControlPlane.Spec.AvailabilityZones) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s order of AZs has changed from %v to %v.",
			key.ControlPlane(&awsControlPlane),
			awsControlPlaneOld.Spec.AvailabilityZones,
			awsControlPlane.Spec.AvailabilityZones),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s order of AZs has changed from %v to %v.",
			key.ControlPlane(&awsControlPlane),
			awsControlPlaneOld.Spec.AvailabilityZones,
			awsControlPlane.Spec.AvailabilityZones),
		)
	}
	return nil
}
func (v *Validator) AZUnique(awsControlPlane infrastructurev1alpha2.AWSControlPlane) error {
	// We always want to select as many distinct AZs as possible
	distinctAZs := countUniqueValues(awsControlPlane.Spec.AvailabilityZones)
	if distinctAZs == len(v.validAvailabilityZones) || distinctAZs == len(awsControlPlane.Spec.AvailabilityZones) {
		return nil
	}
	v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s availability zones %v do not contain maximum amount of distinct AZs. Valid AZs are: %v",
		key.ControlPlane(&awsControlPlane),
		awsControlPlane.Spec.AvailabilityZones,
		v.validAvailabilityZones),
	)
	return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s availability zones %v do not contain maximum amount of distinct AZs. Valid AZs are: %v",
		key.ControlPlane(&awsControlPlane),
		awsControlPlane.Spec.AvailabilityZones,
		v.validAvailabilityZones),
	)
}

func (v *Validator) AZValid(awsControlPlane infrastructurev1alpha2.AWSControlPlane) error {
	if !v.isValidMasterAvailabilityZones(awsControlPlane.Spec.AvailabilityZones) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSControlPlane %s availability zones %v are invalid. Valid AZs are: %v",
			key.ControlPlane(&awsControlPlane),
			awsControlPlane.Spec.AvailabilityZones,
			v.validAvailabilityZones),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s availability zones %v are invalid. Valid AZs are: %v",
			key.ControlPlane(&awsControlPlane),
			awsControlPlane.Spec.AvailabilityZones,
			v.validAvailabilityZones),
		)
	}

	return nil
}

func (v *Validator) ControlPlaneLabelMatch(awsControlPlane infrastructurev1alpha2.AWSControlPlane, g8sControlPlane infrastructurev1alpha2.G8sControlPlane) error {
	if key.ControlPlane(&g8sControlPlane) != key.ControlPlane(&awsControlPlane) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s=%s label does not match with AWSControlPlane %s=%s label for cluster %s",
			label.ControlPlane,
			key.ControlPlane(&g8sControlPlane),
			label.ControlPlane,
			key.ControlPlane(&awsControlPlane),
			key.Cluster(&g8sControlPlane)),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("G8sControlPlane %s=%s label does not match with AWSControlPlane %s=%s label for cluster %s",
			label.ControlPlane,
			key.ControlPlane(&g8sControlPlane),
			label.ControlPlane,
			key.ControlPlane(&awsControlPlane),
			key.Cluster(&g8sControlPlane)),
		)
	}

	return nil
}
func (v *Validator) InstanceTypeValid(awsControlPlane infrastructurev1alpha2.AWSControlPlane) error {
	if !contains(v.validInstanceTypes, awsControlPlane.Spec.InstanceType) {
		return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSControlPlane %s master instance type %v is invalid. Valid instance types are: %v",
			key.ControlPlane(&awsControlPlane),
			awsControlPlane.Spec.InstanceType,
			v.validInstanceTypes),
		)
	}

	return nil
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
	if len(old) <= len(new) {
		for i, o := range old {
			for _, n := range new {
				if o == n && o != new[i] {
					return true
				}
			}
		}
	} else {
		for i, o := range new {
			for _, n := range old {
				if o == n && o != old[i] {
					return true
				}
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
