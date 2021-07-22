package cluster

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/validator"
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
		},
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
	var err error

	// Parse incoming object
	cluster := &capiv1alpha2.Cluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, cluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse awscluster: %v", err)
	}
	err = aws.ValidateOrganizationLabelContainsExistingOrganization(context.Background(), v.k8sClient.CtrlClient(), cluster)
	if err != nil {
		return false, microerror.Mask(err)
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

	capi, err := aws.IsCAPIRelease(cluster)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if capi {
		return true, nil
	}

	if v.isAdmin(request.UserInfo) || v.isInRestrictedGroup(request.UserInfo) {
		err = v.ClusterStatusValid(oldCluster, cluster)
		if err != nil {
			return false, microerror.Mask(err)
		}
		err = v.ClusterLabelKeysValid(oldCluster, cluster)
		if err != nil {
			return false, microerror.Mask(err)
		}
		err = v.ClusterLabelValuesValid(oldCluster, cluster)
		if err != nil {
			return false, microerror.Mask(err)
		}
		err = v.ReleaseVersionValid(oldCluster, cluster)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	return true, nil
}

func (v *Validator) ClusterLabelKeysValid(oldCluster *capiv1alpha2.Cluster, newCluster *capiv1alpha2.Cluster) error {
	return aws.ValidateLabelKeys(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, oldCluster, newCluster)
}

func (v *Validator) ClusterLabelValuesValid(oldCluster *capiv1alpha2.Cluster, newCluster *capiv1alpha2.Cluster) error {
	return aws.ValidateLabelValues(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, oldCluster, newCluster)
}

func (v *Validator) ClusterStatusValid(oldCluster *capiv1alpha2.Cluster, newCluster *capiv1alpha2.Cluster) error {
	var err error

	if key.Release(newCluster) == key.Release(oldCluster) {
		return nil
	}
	// Retrieve the `AWSCluster` CR.
	awsCluster, err := aws.FetchAWSCluster(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, newCluster)
	if err != nil {
		return microerror.Mask(err)
	}
	if !v.isTransitioned(awsCluster.GetCommonClusterStatus()) {
		return microerror.Maskf(notAllowedError, "Cluster %v can not be upgraded at the present moment because it has not transitioned yet.",
			newCluster.GetName(),
		)
	}

	return nil
}

func (v *Validator) ReleaseVersionValid(oldCluster *capiv1alpha2.Cluster, newCluster *capiv1alpha2.Cluster) error {
	var err error

	if key.Release(newCluster) == key.Release(oldCluster) {
		return nil
	}
	releaseVersion, err := aws.ReleaseVersion(newCluster, []mutator.PatchOperation{})
	if err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}
	oldReleaseVersion, err := aws.ReleaseVersion(oldCluster, []mutator.PatchOperation{})
	if err != nil {
		return microerror.Maskf(parsingFailedError, "unable to parse release version from Cluster")
	}
	if releaseVersion.Major < oldReleaseVersion.Major {
		return microerror.Maskf(notAllowedError, "Upgrade from %v to %v is a major downgrade and is not supported.",
			oldReleaseVersion.String(),
			releaseVersion.String())
	}
	if releaseVersion.Major > oldReleaseVersion.Major+1 {
		return microerror.Maskf(notAllowedError, "Upgrade from %v to %v skips major release versions and is not supported.",
			oldReleaseVersion.String(),
			releaseVersion.String())
	}
	// Retrieve the `Release` CR.
	release, err := aws.FetchRelease(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, releaseVersion)
	if err != nil {
		return microerror.Mask(err)
	}
	if release.Spec.State == "deprecated" {
		return microerror.Maskf(notAllowedError, "Release %v is deprecated.", release.GetName())
	}

	return nil
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

func (v *Validator) isTransitioned(status infrastructurev1alpha2.CommonClusterStatus) bool {
	condition := status.LatestCondition()
	return condition == infrastructurev1alpha2.ClusterStatusConditionCreated || condition == infrastructurev1alpha2.ClusterStatusConditionUpdated
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "awscluster"
}
