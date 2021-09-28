package g8scontrolplane

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v3/config"
	aws "github.com/giantswarm/aws-admission-controller/v3/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/validator"
)

type Validator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewValidator(config config.Config) (*Validator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	validator := &Validator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
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

func (v *Validator) ValidateCreate(request *admissionv1.AdmissionRequest) (bool, error) {
	var g8sControlPlane infrastructurev1alpha3.G8sControlPlane
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &g8sControlPlane); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscontrol plane: %v", err)
	}

	err = aws.ValidateOrgNamespace(&g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = aws.ValidateOperatorVersion(&g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = aws.ValidateOrganizationLabelContainsExistingOrganization(context.Background(), v.k8sClient.CtrlClient(), &g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.ReplicaCount(g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = v.ReplicaAZMatch(g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) ValidateUpdate(request *admissionv1.AdmissionRequest) (bool, error) {
	var g8sControlPlane infrastructurev1alpha3.G8sControlPlane
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &g8sControlPlane); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscontrol plane: %v", err)
	}

	err = v.ReplicaCount(g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = v.ReplicaAZMatch(g8sControlPlane)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) ReplicaAZMatch(g8sControlPlane infrastructurev1alpha3.G8sControlPlane) error {
	var err error

	// Retrieve the `AWSControlPlane` CR related to this object.
	awsControlPlane, err := aws.FetchAWSControlPlane(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, &g8sControlPlane)
	// Note that while we do log the error, we don't fail if the AWSControlPlane doesn't exist yet. That is okay because the order of CR creation can vary.
	if aws.IsNotFound(err) {
		v.Log("level", "debug", "message", fmt.Sprintf("No AWSControlPlane %s could be found: %v", g8sControlPlane.GetName(), err))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	if g8sControlPlane.Spec.Replicas != len(awsControlPlane.Spec.AvailabilityZones) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s with %v replicas does not match AWSControlPlane %s with %v availability zones %s",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			key.ControlPlane(awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			awsControlPlane.Spec.AvailabilityZones),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("G8sControlPlane %s with %v replicas does not match AWSControlPlane %s with %v availability zones %s",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			key.ControlPlane(awsControlPlane),
			len(awsControlPlane.Spec.AvailabilityZones),
			awsControlPlane.Spec.AvailabilityZones),
		)
	}

	return nil
}
func (v *Validator) ReplicaCount(g8sControlPlane infrastructurev1alpha3.G8sControlPlane) error {
	if !aws.IsValidMasterReplicas(g8sControlPlane.Spec.Replicas) {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("G8sControlPlane %s has an invalid count of %v replicas. Valid replica counts are: %v",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			aws.ValidMasterReplicas()),
		)
		return microerror.Maskf(notAllowedError, fmt.Sprintf("G8sControlPlane %s has an invalid count of %v replicas. Valid replica counts are: %v",
			key.ControlPlane(&g8sControlPlane),
			g8sControlPlane.Spec.Replicas,
			aws.ValidMasterReplicas()),
		)
	}

	return nil
}

func (m *Validator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "g8scontrolplane"
}
