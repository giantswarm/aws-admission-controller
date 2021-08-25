package machinedeployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/aws-admission-controller/v3/config"
	aws "github.com/giantswarm/aws-admission-controller/v3/pkg/aws/v1alpha3"
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
	if request.Operation == admissionv1.Create {
		return v.ValidateCreate(request)
	}
	return true, nil
}

func (v *Validator) ValidateCreate(request *admissionv1.AdmissionRequest) (bool, error) {
	var err error

	var machineDeployment capiv1alpha3.MachineDeployment
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &machineDeployment); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse machinedeployment: %v", err)
	}
	capi, err := aws.IsCAPIRelease(&machineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if capi {
		return true, nil
	}

	err = v.ValidateCluster(machineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = aws.ValidateOrganizationLabelContainsExistingOrganization(context.Background(), v.k8sClient.CtrlClient(), &machineDeployment)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) ValidateCluster(machineDeployment capiv1alpha3.MachineDeployment) error {
	var err error

	// Retrieve the `Cluster` CR related to this object.
	cluster, err := aws.FetchCluster(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, &machineDeployment)
	if err != nil {
		return microerror.Mask(err)
	}
	// make sure the cluster is not deleted
	if cluster.DeletionTimestamp != nil {
		return microerror.Maskf(notAllowedError, fmt.Sprintf("MachineDeployment could not be created because Cluster '%s' is in deleting state.",
			cluster.Name),
		)
	}
	return nil
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "machinedeployment"
}
