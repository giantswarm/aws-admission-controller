package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/blang/semver/v4"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/aws-admission-controller/v4/config"
	aws "github.com/giantswarm/aws-admission-controller/v4/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/validator"
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
	cluster := &capi.Cluster{}
	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, cluster); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse cluster: %v", err)
	}
	err = aws.ValidateOrgNamespace(cluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = aws.ValidateOperatorVersion(cluster)
	if err != nil {
		return false, microerror.Mask(err)
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
	cluster := &capi.Cluster{}
	oldCluster := &capi.Cluster{}
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

	err = v.Cilium(cluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.ClusterAnnotationUpgradeTimeIsValid(cluster, oldCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	err = v.ClusterAnnotationUpgradeReleaseIsValid(cluster)
	if err != nil {
		return false, microerror.Mask(err)
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

func (v *Validator) Cilium(cluster *capi.Cluster) error {
	release, err := semver.New(key.Release(cluster))
	if err != nil {
		return err
	}

	if !aws.IsCiliumRelease(release) {
		return nil
	}

	// Retrieve the `AWSCluster` CR.
	awsCluster, err := aws.FetchAWSCluster(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, cluster)
	if err != nil {
		return err
	}
	_, ok := awsCluster.Annotations["cilium.giantswarm.io/pod-cidr"]
	if !ok {
		return microerror.Maskf(notAllowedError,
			fmt.Sprint("The annotation `cilium.giantswarm.io/pod-cidr` has to be set on AWSCluster CR before upgrading to AWS release v18 or higher."),
		)
	}

	return nil
}

func (v *Validator) ClusterAnnotationUpgradeTimeIsValid(cluster *capi.Cluster, oldCluster *capi.Cluster) error {
	if updateTime, ok := cluster.GetAnnotations()[annotation.UpdateScheduleTargetTime]; ok {
		if updateTimeOld, ok := oldCluster.GetAnnotations()[annotation.UpdateScheduleTargetTime]; ok {
			if updateTime == updateTimeOld {
				return nil
			}
		}
		v.logger.Log("level", "debug", "message", fmt.Sprintf("upgrade time is set to %s", updateTime))
		if !UpgradeScheduleTimeIsValid(updateTime) {
			v.logger.Log("level", "error", "message", "upgrade time is not valid")
			return microerror.Maskf(notAllowedError,
				fmt.Sprintf("Cluster annotation '%s' value '%s' is not valid. Value must be in RFC822 format and UTC time zone (e.g. 30 Jan 21 15:04 UTC) and should be a date 16 mins - 6months in the future.",
					annotation.UpdateScheduleTargetTime,
					updateTime),
			)
		}
	}
	return nil
}

func UpgradeScheduleTimeIsValid(updateTime string) bool {
	// parse time
	t, err := time.Parse(time.RFC822, updateTime)
	if err != nil {
		return false
	}
	// check whether it is UTC
	if t.Location().String() != "UTC" {
		return false
	}
	// time already passed or is less than 16 minutes in the future
	if t.Before(time.Now().UTC().Add(16 * time.Minute)) {
		return false
	}
	// time is 6 months or more in the future (6 months are 4380 hours)
	if t.Sub(time.Now().UTC()) > 4380*time.Hour {
		return false
	}
	return true
}

func (v *Validator) ClusterAnnotationUpgradeReleaseIsValid(cluster *capi.Cluster) error {
	if targetRelease, ok := cluster.GetAnnotations()[annotation.UpdateScheduleTargetRelease]; ok {
		v.logger.Log("level", "debug", "message", fmt.Sprintf("upgrade release is set to %s", targetRelease))
		err := v.UpgradeScheduleReleaseIsValid(targetRelease, key.Release(cluster))
		if err != nil {
			v.logger.Log("level", "error", "message", err)
			return microerror.Maskf(notAllowedError,
				fmt.Sprintf("Cluster annotation '%s' value '%s' is not valid. Value must be an existing giant swarm release version above the current release version %s and must not have a v prefix. %v",
					annotation.UpdateScheduleTargetTime,
					targetRelease,
					key.Release(cluster),
					err),
			)
		}
	}
	return nil
}

func (v *Validator) UpgradeScheduleReleaseIsValid(targetRelease string, currentRelease string) error {
	// parse target version
	t, err := semver.New(targetRelease)
	if err != nil {
		return microerror.Mask(err)
	}
	// check if the release exists
	_, err = aws.FetchRelease(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, t)
	if err != nil {
		return microerror.Mask(err)
	}
	// parse current version
	c, err := semver.New(currentRelease)
	if err != nil {
		return microerror.Mask(err)
	}
	// check if target is higher than the current release
	if t.LE(*c) {
		return microerror.Maskf(notAllowedError, "Upgrade target release version has to be above current release version.")
	}
	return nil
}

func (v *Validator) ClusterLabelKeysValid(oldCluster *capi.Cluster, newCluster *capi.Cluster) error {
	return aws.ValidateLabelKeys(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, oldCluster, newCluster)
}

func (v *Validator) ClusterLabelValuesValid(oldCluster *capi.Cluster, newCluster *capi.Cluster) error {
	return aws.ValidateLabelValues(&aws.Handler{K8sClient: v.k8sClient, Logger: v.logger}, oldCluster, newCluster)
}

func (v *Validator) ClusterStatusValid(oldCluster *capi.Cluster, newCluster *capi.Cluster) error {
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

func (v *Validator) ReleaseVersionValid(oldCluster *capi.Cluster, newCluster *capi.Cluster) error {
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

func (v *Validator) isTransitioned(status infrastructurev1alpha3.CommonClusterStatus) bool {
	condition := status.LatestCondition()
	return condition == infrastructurev1alpha3.ClusterStatusConditionCreated || condition == infrastructurev1alpha3.ClusterStatusConditionUpdated
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "cluster"
}
