package cluster

import (
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
)

type Validator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	restrictedGroups []string
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

		restrictedGroups: []string{
			config.AdminGroup,
			config.AllTargetGroup,
		},
	}

	return v, nil
}

func (v *Validator) Validate(request *admissionv1.AdmissionRequest) (bool, error) {
	if request.Operation == admissionv1.Update {
		return v.ValidateUpdate(request)
	}
	return true, nil
}

func (v *Validator) ValidateUpdate(request *admissionv1.AdmissionRequest) (bool, error) {
	var err error

	// Parse incoming object
	cluster := &capiv1alpha2.Cluster{}
	oldCluster := &capiv1alpha2.Cluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, cluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse Cluster: %v", err)
	}
	if _, _, err := mutator.Deserializer.Decode(request.OldObject.Raw, nil, oldCluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse old Cluster: %v", err)
	}

	if v.isAdmin(request.UserInfo) || v.isInRestrictedGroup(request.UserInfo) {
		err = v.ClusterLabelValuesValid(oldCluster, cluster)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	return true, nil
}

func (v *Validator) ClusterLabelValuesValid(oldCluster *capiv1alpha2.Cluster, newCluster *capiv1alpha2.Cluster) error {
	return aws.ValidateLabelValues(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, oldCluster, newCluster)
}

func (v *Validator) isAdmin(userInfo authenticationv1.UserInfo) bool {
	for _, u := range aws.ValidLabelAdmins() {
		if u == userInfo.Username {
			return true
		}
	}
	return false
}

func (v *Validator) isInRestrictedGroup(userInfo authenticationv1.UserInfo) bool {
	for _, r := range v.restrictedGroups {
		for _, u := range userInfo.Groups {
			if r == u {
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
	return "awscluster"
}
