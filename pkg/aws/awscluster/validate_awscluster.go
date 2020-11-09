package awscluster

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/validator"
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
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsCluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscluster: %v", err)
	}

	err = v.AWSClusterAnnotationMaxBatchSizeIsValid(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationPauseTimeIsValid(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) AWSClusterAnnotationMaxBatchSizeIsValid(awsCluster infrastructurev1alpha2.AWSCluster) error {
	if maxBatchSize, ok := awsCluster.GetAnnotations()[aws.AnnotationUpdateMaxBatchSize]; ok {
		if !aws.MaxBatchSizeIsValid(maxBatchSize) {
			v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Allowed value is either integer bigger than zero or decimal number between 0 and 1.0 defining percentage of nodes",
				aws.AnnotationUpdateMaxBatchSize,
				maxBatchSize),
			)
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Allowed value is either integer bigger than zero or decimal number between 0 and 1.0 defining percentage of nodes",
				aws.AnnotationUpdateMaxBatchSize,
				maxBatchSize),
			)
		}
	}
	return nil
}

func (v *Validator) AWSClusterAnnotationPauseTimeIsValid(awsCluster infrastructurev1alpha2.AWSCluster) error {
	if maxBatchSize, ok := awsCluster.GetAnnotations()[aws.AnnotationUpdatePauseTime]; ok {
		if !aws.PauseTimeIsValid(maxBatchSize) {
			v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be in ISO 8601 duration format and cannot be bigger than 1 hour.",
				aws.AnnotationUpdatePauseTime,
				maxBatchSize),
			)
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be in ISO 8601 duration format and cannot be bigger than 1 hour.",
				aws.AnnotationUpdatePauseTime,
				maxBatchSize),
			)
		}
	}
	return nil
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "awscluster"
}
