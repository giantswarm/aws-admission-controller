package awscluster

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/validator"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
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

	v := &Validator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return v, nil
}

func (v *Validator) Validate(request *admissionv1.AdmissionRequest) (bool, error) {
	var awsCluster infrastructurev1alpha2.AWSCluster
	///var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsCluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscluster: %v", err)
	}

	// TODO

	//err = v.MachineDeploymentLabelMatch(awsMachineDeployment)
	//if err != nil {
	//	return false, microerror.Mask(err)

	//	}

	return true, nil
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "awscluster"
}
