package awscluster

import (
	"context"
	"fmt"
	"strings"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-admission-controller/v4/config"
	aws "github.com/giantswarm/aws-admission-controller/v4/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/validator"
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
	if request.Operation == admissionv1.Create {
		return v.ValidateCreate(request)
	}
	if request.Operation == admissionv1.Update {
		return v.ValidateUpdate(request)
	}
	return true, nil
}

func (v *Validator) ValidateCreate(request *admissionv1.AdmissionRequest) (bool, error) {
	var awsCluster infrastructurev1alpha3.AWSCluster
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsCluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscluster: %v", err)
	}

	err = aws.ValidateOrgNamespace(&awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = aws.ValidateOperatorVersion(&awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = aws.ValidateOrganizationLabelContainsExistingOrganization(context.Background(), v.k8sClient.CtrlClient(), &awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationMaxBatchSizeIsValid(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationPauseTimeIsValid(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationCNIMinimumIPTarget(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationCNIWarmIPTarget(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = v.AWSClusterAnnotationNodeTerminateUnhealthy(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationCNIPrefix(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) ValidateUpdate(request *admissionv1.AdmissionRequest) (bool, error) {
	var awsCluster infrastructurev1alpha3.AWSCluster
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &awsCluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscluster: %v", err)
	}

	err = aws.ValidateOrganizationLabelContainsExistingOrganization(context.Background(), v.k8sClient.CtrlClient(), &awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationMaxBatchSizeIsValid(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationPauseTimeIsValid(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationCNIMinimumIPTarget(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.AWSClusterAnnotationCNIWarmIPTarget(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = v.AWSClusterAnnotationNodeTerminateUnhealthy(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}
	err = v.AWSClusterAnnotationCNIPrefix(awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) AWSClusterAnnotationCNIMinimumIPTarget(awsCluster infrastructurev1alpha3.AWSCluster) error {
	if cniMinimumIPTarget, ok := awsCluster.GetAnnotations()[annotation.AWSCNIMinimumIPTarget]; ok {
		if !aws.IsIntegerGreaterThanZero(cniMinimumIPTarget) {
			v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be a integer greater than zero.",
				annotation.AWSCNIMinimumIPTarget,
				cniMinimumIPTarget),
			)
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be a integer greater than zero.",
				annotation.AWSCNIMinimumIPTarget,
				cniMinimumIPTarget),
			)
		}
	}
	return nil
}

func (v *Validator) AWSClusterAnnotationCNIWarmIPTarget(awsCluster infrastructurev1alpha3.AWSCluster) error {
	if cniWarmIPTarget, ok := awsCluster.GetAnnotations()[annotation.AWSCNIWarmIPTarget]; ok {
		if !aws.IsIntegerGreaterThanZero(cniWarmIPTarget) {
			v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be a integer greater than zero.",
				annotation.AWSCNIWarmIPTarget,
				cniWarmIPTarget),
			)
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be a integer greater than zero.",
				annotation.AWSCNIWarmIPTarget,
				cniWarmIPTarget),
			)
		}
	}
	return nil
}

func (v *Validator) AWSClusterAnnotationMaxBatchSizeIsValid(awsCluster infrastructurev1alpha3.AWSCluster) error {
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

func (v *Validator) AWSClusterAnnotationPauseTimeIsValid(awsCluster infrastructurev1alpha3.AWSCluster) error {
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

func (v *Validator) AWSClusterAnnotationNodeTerminateUnhealthy(awsCluster infrastructurev1alpha3.AWSCluster) error {
	if terminateUnhealthy, ok := awsCluster.GetAnnotations()[annotation.NodeTerminateUnhealthy]; ok {
		if !(terminateUnhealthy == stringTrue || terminateUnhealthy == stringFalse) {
			v.logger.Log("level", "debug", "message", fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be either '\"true\"' or '\"false\"'.",
				annotation.NodeTerminateUnhealthy,
				terminateUnhealthy),
			)
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSCluster annotation '%s' value '%s' is not valid. Value must be either '\"true\"' or '\"false\"'.",
				annotation.NodeTerminateUnhealthy,
				terminateUnhealthy),
			)
		}
	}
	return nil
}

func (v *Validator) AWSClusterAnnotationCNIPrefix(awsCluster infrastructurev1alpha3.AWSCluster) error {
	if _, ok := awsCluster.GetAnnotations()[annotation.AWSCNIPrefixDelegation]; ok {
		if strings.Contains(awsCluster.Spec.Provider.Region, "cn-") {
			return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSCluster annotation '%s' is not allowed in region 'China'.",
				annotation.AWSCNIPrefixDelegation),
			)
		}
	}
	return nil
}

func (v *Validator) AWSClusterExists(obj metav1.Object) error {
	// Parse existing AWS clusters
	awsClusters := &infrastructurev1alpha3.AWSClusterList{}
	err := v.k8sClient.CtrlClient().List(context.Background(), awsClusters)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(awsClusters.Items) > 0 {
		for _, cluster := range awsClusters.Items {
			if obj.GetName() == cluster.Name {
				return microerror.Maskf(notAllowedError, fmt.Sprintf("AWSCluster %s/%s already exists", cluster.Namespace, cluster.Name))
			}
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
