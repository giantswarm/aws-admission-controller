package cluster

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	unittest "github.com/giantswarm/aws-admission-controller/v4/pkg/unittest/v1alpha3"
)

func Test_UpgradeReleaseIsValid(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		oldValue     string
		noAnnotation bool
		valid        bool
	}{
		{
			name:     "case 0: minor upgrade",
			oldValue: "14.0.0",
			value:    "14.1.0",
			valid:    true,
		},
		{
			name:     "case 1: major upgrade",
			oldValue: "14.1.0",
			value:    "15.1.0",
			valid:    true,
		},
		{
			name:     "case 2: no value",
			oldValue: "14.1.0",
			value:    "",
			valid:    false,
		},
		{
			name:         "case 3: no annotation",
			oldValue:     "14.1.0",
			noAnnotation: true,
			valid:        true,
		},
		{
			name:     "case 4: no upgrade",
			oldValue: "14.1.0",
			value:    "14.1.0",
			valid:    false,
		},
		{
			name:     "case 5: downgrade",
			oldValue: "15.1.0",
			value:    "14.1.0",
			valid:    false,
		},
		{
			name:     "case 6: release does not exist",
			oldValue: "14.1.0",
			value:    "14.11.0",
			valid:    false,
		},
		{
			name:     "case 7: v prefix",
			oldValue: "14.1.0",
			value:    "v15.1.0",
			valid:    false,
		},
		{
			name:     "case 8: invalid format",
			oldValue: "14.1.0",
			value:    "15-1-0",
			valid:    false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := &Validator{
				k8sClient: unittest.FakeK8sClient(),
				logger:    microloggertest.New(),
			}
			cluster := unittest.DefaultCluster()
			cluster.SetLabels(map[string]string{label.ReleaseVersion: tc.oldValue})
			if !tc.noAnnotation {
				cluster.SetAnnotations(map[string]string{annotation.UpdateScheduleTargetRelease: tc.value})
			}

			// create releases for testing
			releases := []unittest.ReleaseData{
				{
					Name: "v14.1.0",
				},
				{
					Name: "v15.1.0",
				},
			}
			for _, r := range releases {
				release := unittest.DefaultRelease()
				release.SetName(r.Name)
				err := v.k8sClient.CtrlClient().Create(context.Background(), &release)
				if err != nil {
					t.Fatal(err)
				}
			}
			// check if the result is as expected
			err := v.ClusterAnnotationUpgradeReleaseIsValid(cluster)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func Test_UpgradeTimeIsValid(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		noAnnotation bool
		noChange     bool
		valid        bool
	}{
		{
			name:  "case 0: 2 hours from now in RFC822 format",
			value: time.Now().UTC().Add(2 * time.Hour).Format(time.RFC822),
			valid: true,
		},
		{
			name:  "case 1: 2 hours from now in RFC850 format",
			value: time.Now().UTC().Add(2 * time.Hour).Format(time.RFC850),
			valid: false,
		},
		{
			name:  "case 2: no value",
			value: "",
			valid: false,
		},
		{
			name:         "case 3: no annotation",
			noAnnotation: true,
			valid:        true,
		},
		{
			name:  "case 4: More than 6 months later",
			value: time.Now().UTC().Add(4381 * time.Hour).Format(time.RFC822),
			valid: false,
		},
		{
			name:  "case 5: 2 minutes before",
			value: time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC850),
			valid: false,
		},
		{
			name:  "case 6: not UTC",
			value: strings.Replace(time.Now().UTC().Add(-2*time.Minute).Format(time.RFC850), "UTC", "CET", 1),
			valid: false,
		},
		{
			name:  "case 7: 15 minutes from now",
			value: time.Now().UTC().Add(15 * time.Minute).Format(time.RFC822),
			valid: false,
		},
		{
			name:     "case 8: 15 minutes from now but no change to annotation.",
			value:    time.Now().UTC().Add(15 * time.Minute).Format(time.RFC822),
			noChange: true,
			valid:    true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := &Validator{
				k8sClient: unittest.FakeK8sClient(),
				logger:    microloggertest.New(),
			}
			cluster := unittest.DefaultCluster()
			if !tc.noAnnotation {
				cluster.SetAnnotations(map[string]string{annotation.UpdateScheduleTargetTime: tc.value})
			}
			oldCluster := unittest.DefaultCluster()
			if tc.noChange {
				oldCluster.SetAnnotations(map[string]string{annotation.UpdateScheduleTargetTime: tc.value})
			}
			// check if the result is as expected
			err := v.ClusterAnnotationUpgradeTimeIsValid(cluster, oldCluster)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func TestValidateReleaseVersion(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		oldReleaseVersion string
		newReleaseVersion string
		valid             bool
	}{
		{
			// Version unchanged
			name: "case 0",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "3.0.0",
			valid:             true,
		},
		{
			// version changed to valid release
			name: "case 1",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "4.0.0",
			valid:             true,
		},
		{
			// version changed to deprecated release
			name: "case 2",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "3.2.0",
			valid:             false,
		},
		{
			// version changed to invalid release
			name: "case 3",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "3.3.0",
			valid:             false,
		},
		{
			// version changed with major skip
			name: "case 4",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "5.0.0",
			valid:             false,
		},
		{
			// version changed to older release
			name: "case 5",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "2.0.0",
			valid:             false,
		},
		{
			// version changed with multiple major skips
			name: "case 6",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "7.0.0",
			valid:             false,
		},
		{
			// version changed to older minor release
			name: "case 7",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "2.9.0",
			valid:             false,
		},
		{
			// version changed to older minor release
			name: "case 8",
			ctx:  context.Background(),

			oldReleaseVersion: "3.2.0",
			newReleaseVersion: "3.1.0",
			valid:             true,
		},
		{
			// version changed to older minor release
			name: "case 9",
			ctx:  context.Background(),

			oldReleaseVersion: "3.4.1",
			newReleaseVersion: "3.1.0",
			valid:             true,
		},
		{
			// version changed to older patch release
			name: "case 10",
			ctx:  context.Background(),

			oldReleaseVersion: "3.2.2",
			newReleaseVersion: "3.2.1",
			valid:             true,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			handle := &Validator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			// create releases for testing
			releases := []unittest.ReleaseData{
				{
					Name:  "v5.0.0",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v4.0.0",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v3.4.1",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v3.2.2",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v3.2.1",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v3.2.0",
					State: releasev1alpha1.StateDeprecated,
				},
				{
					Name:  "v3.1.0",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v2.0.0",
					State: releasev1alpha1.StateActive,
				},
			}
			for _, r := range releases {
				release := unittest.DefaultRelease()
				release.SetName(r.Name)
				release.Spec.State = r.State
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, &release)
				if err != nil {
					t.Fatal(err)
				}
			}

			// create old and new object with release version labels
			oldObject := unittest.DefaultCluster()
			oldLabels := unittest.DefaultLabels()
			oldLabels[label.ReleaseVersion] = tc.oldReleaseVersion
			oldObject.SetLabels(oldLabels)

			newObject := unittest.DefaultCluster()
			newLabels := unittest.DefaultLabels()
			newLabels[label.ReleaseVersion] = tc.newReleaseVersion
			newObject.SetLabels(newLabels)

			// check if the result is as expected
			err = handle.ReleaseVersionValid(oldObject, newObject)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func TestValidClusterStatus(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		oldReleaseVersion string
		newReleaseVersion string
		conditions        []infrastructurev1alpha3.CommonClusterStatusCondition

		valid bool
	}{
		{
			// no upgrade
			name: "case 0",
			ctx:  context.Background(),

			conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
				{LastTransitionTime: metav1.NewTime(time.Now()),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreating},
			},
			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "3.0.0",
			valid:             true,
		},
		{
			// Cluster is creating
			name: "case 1",
			ctx:  context.Background(),

			conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
				{LastTransitionTime: metav1.NewTime(time.Now()),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreating},
			},
			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "4.0.0",
			valid:             false,
		},
		{
			// Cluster is created
			name: "case 2",
			ctx:  context.Background(),

			conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
				{LastTransitionTime: metav1.NewTime(time.Now()),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreated},
				{LastTransitionTime: metav1.NewTime(time.Now().Add(-15 * time.Minute)),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreating},
			},
			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "4.0.0",
			valid:             true,
		},
		{
			// Cluster is updating
			name: "case 3",
			ctx:  context.Background(),

			conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
				{LastTransitionTime: metav1.NewTime(time.Now()),
					Condition: infrastructurev1alpha3.ClusterStatusConditionUpdating},
				{LastTransitionTime: metav1.NewTime(time.Now().Add(-15 * time.Minute)),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreated},
				{LastTransitionTime: metav1.NewTime(time.Now().Add(-30 * time.Minute)),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreating},
			},
			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "4.0.0",
			valid:             false,
		},
		{
			// Cluster is updated
			name: "case 4",
			ctx:  context.Background(),

			conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
				{LastTransitionTime: metav1.NewTime(time.Now()),
					Condition: infrastructurev1alpha3.ClusterStatusConditionUpdated},
				{LastTransitionTime: metav1.NewTime(time.Now().Add(-15 * time.Minute)),
					Condition: infrastructurev1alpha3.ClusterStatusConditionUpdating},
				{LastTransitionTime: metav1.NewTime(time.Now().Add(-30 * time.Minute)),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreated},
				{LastTransitionTime: metav1.NewTime(time.Now().Add(-60 * time.Minute)),
					Condition: infrastructurev1alpha3.ClusterStatusConditionCreating},
			},
			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "4.0.0",
			valid:             true,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			handle := &Validator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			awsCluster := unittest.DefaultAWSCluster()
			awsCluster.Status.Cluster.Conditions = tc.conditions
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsCluster)
			if err != nil {
				t.Fatal(err)
			}

			// create old and new object with release version labels
			oldObject := unittest.DefaultCluster()
			oldLabels := unittest.DefaultLabels()
			oldLabels[label.ReleaseVersion] = tc.oldReleaseVersion
			oldObject.SetLabels(oldLabels)

			newObject := unittest.DefaultCluster()
			newLabels := unittest.DefaultLabels()
			newLabels[label.ReleaseVersion] = tc.newReleaseVersion
			newObject.SetLabels(newLabels)

			// check if the result is as expected
			err = handle.ClusterStatusValid(oldObject, newObject)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func TestCilium(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentRelease string
		targetRelease  string
		ciliumCidr     string
		ipamCidrBlock  string
		podCidrBlock   string
		err            error
	}{
		{
			// CNI CIDR is not allowed too small.
			name: "case 0",
			ctx:  context.Background(),

			currentRelease: "17.4.1",
			targetRelease:  "18.0.0",
			ciliumCidr:     "10.0.0.0/25",
			ipamCidrBlock:  "10.5.0.0/16",
			podCidrBlock:   "10.4.0.0/16",
			err:            notAllowedError,
		},
		{
			// CNI CIDR is not allowed, overlapping CIDR's.
			name: "case 1",
			ctx:  context.Background(),

			currentRelease: "17.4.1",
			targetRelease:  "18.0.0",
			ciliumCidr:     "10.0.0.0/8",
			ipamCidrBlock:  "10.5.0.0/16",
			podCidrBlock:   "10.0.0.0/16",
			err:            notAllowedError,
		},
		{
			// CNI CIDR is allowed, no overlapping CIDR's.
			name: "case 2",
			ctx:  context.Background(),

			currentRelease: "17.4.1",
			targetRelease:  "18.0.0",
			ciliumCidr:     "10.0.0.0/16",
			ipamCidrBlock:  "10.1.0.0/16",
			podCidrBlock:   "10.2.0.0/16",
			err:            nil,
		},
		{
			// Not an upgrade
			name: "case 3",
			ctx:  context.Background(),

			currentRelease: "17.4.1",
			targetRelease:  "17.5.0",
			ciliumCidr:     "",
			ipamCidrBlock:  "10.5.0.0/16",
			podCidrBlock:   "10.0.0.0/16",
			err:            nil,
		},
		{
			// Not an upgrade
			name: "case 4",
			ctx:  context.Background(),

			currentRelease: "18.4.1",
			targetRelease:  "18.4.1",
			ciliumCidr:     "",
			ipamCidrBlock:  "10.5.0.0/16",
			podCidrBlock:   "10.0.0.0/16",
			err:            nil,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),

				ipamCidrBlock: tc.ipamCidrBlock,
			}

			// run admission request to default AWSCluster Pod CIDR
			awsCluster := unittest.DefaultAWSCluster()
			awsCluster.Spec.Provider.Pods.CIDRBlock = tc.podCidrBlock
			err = fakeK8sClient.CtrlClient().Create(context.Background(), awsCluster)
			if err != nil {
				t.Fatal(err)
			}

			oldCluster := unittest.DefaultCluster()
			oldCluster.Labels[label.ReleaseVersion] = tc.currentRelease

			cluster := unittest.DefaultCluster()
			// set CNI prefix annotation
			if tc.ciliumCidr != "" {
				cluster.SetAnnotations(map[string]string{
					annotation.CiliumPodCidr: tc.ciliumCidr,
				})
			}
			cluster.Labels[label.ReleaseVersion] = tc.targetRelease

			err = validate.Cilium(cluster, oldCluster)
			if microerror.Cause(err) != tc.err {
				t.Fatal(err)
			}
		})
	}
}

func Test_ClusterAlreadyExists(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
	}{
		{
			name:  "case 0: cluster does not exists",
			valid: true,
		},
		{
			name:  "case 1: cluster already exists",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := &Validator{
				k8sClient: unittest.FakeK8sClient(),
				logger:    microloggertest.New(),
			}
			cluster := unittest.DefaultCluster()

			if !tc.valid {
				// create a cluster in giantswarm namespace
				cluster.Namespace = "giantswarm"
				err := v.k8sClient.CtrlClient().Create(context.TODO(), cluster)
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
			}

			// check if the result is as expected
			err := v.ClusterExists(cluster)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func TestValidator_EnsureGitopsPaused(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentRelease string
		targetRelease  string
		kustomization  *v1beta2.Kustomization
		labels         map[string]string
		err            error
	}{
		{
			name: "case 0: no upgrade to v19",
			ctx:  context.Background(),

			currentRelease: "17.4.1",
			targetRelease:  "18.0.0",
			kustomization:  nil,
			labels:         nil,
			err:            nil,
		},
		{
			name: "case 1: already in v19",
			ctx:  context.Background(),

			currentRelease: "19.0.0",
			targetRelease:  "19.0.1",
			kustomization:  nil,
			labels:         nil,
			err:            nil,
		},
		{
			name: "case 2: upgrade to v19, no flux",
			ctx:  context.Background(),

			currentRelease: "18.3.0",
			targetRelease:  "19.0.0-beta1",
			kustomization:  nil,
			labels:         nil,
			err:            nil,
		},
		{
			name: "case 3: upgrade to v19, flux but suspended",
			ctx:  context.Background(),

			currentRelease: "18.3.0",
			targetRelease:  "19.0.0-beta1",
			kustomization: &v1beta2.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kust1",
					Namespace: "default",
				},
				Spec: v1beta2.KustomizationSpec{
					Suspend: true,
				},
			},
			labels: map[string]string{
				"kustomize.toolkit.fluxcd.io/name":      "kust1",
				"kustomize.toolkit.fluxcd.io/namespace": "default",
			},
			err: nil,
		},
		{
			name: "case 4: upgrade to v19, flux not suspended",
			ctx:  context.Background(),

			currentRelease: "18.3.0",
			targetRelease:  "19.0.0-beta1",
			kustomization: &v1beta2.Kustomization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kust1",
					Namespace: "default",
				},
				Spec: v1beta2.KustomizationSpec{
					Suspend: false,
				},
			},
			labels: map[string]string{
				"kustomize.toolkit.fluxcd.io/name":      "kust1",
				"kustomize.toolkit.fluxcd.io/namespace": "default",
			},
			err: notAllowedError,
		},
		{
			name: "case 5: upgrade to v19, flux labels present but no kustomization exists",
			ctx:  context.Background(),

			currentRelease: "18.3.0",
			targetRelease:  "19.0.0-beta1",
			kustomization:  nil,
			labels: map[string]string{
				"kustomize.toolkit.fluxcd.io/name":      "kust1",
				"kustomize.toolkit.fluxcd.io/namespace": "default",
			},
			err: nil,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()

			// Create kustomization CR
			if tc.kustomization != nil {
				err = fakeK8sClient.CtrlClient().Create(context.Background(), tc.kustomization)
				if err != nil {
					t.Fatal(err)
				}
			}

			oldCluster := unittest.DefaultCluster()
			oldCluster.Labels[label.ReleaseVersion] = tc.currentRelease

			cluster := unittest.DefaultCluster()
			cluster.Labels[label.ReleaseVersion] = tc.targetRelease
			for k, v := range tc.labels {
				cluster.Labels[k] = v
			}

			validate := &Validator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			err = validate.EnsureGitopsPaused(cluster, oldCluster)
			if microerror.Cause(err) != tc.err {
				t.Fatal(err)
			}
		})
	}
}

func Test_CiliumIpamMode(t *testing.T) {
	testCases := []struct {
		name        string
		annotations map[string]string
		err         error
	}{
		{
			name:        "case 0: nil annotations",
			annotations: nil,
			err:         nil,
		},
		{
			name:        "case 1: empty map of annotations",
			annotations: map[string]string{},
			err:         nil,
		},
		{
			name: "case 2: some annotations but not the one we care about",
			annotations: map[string]string{
				"giantswarm.io/test": "test",
			},
			err: nil,
		},
		{
			name: "case 3: annotation with valid value 'eni'",
			annotations: map[string]string{
				"cilium.giantswarm.io/ipam-mode": "eni",
			},
			err: nil,
		},
		{
			name: "case 4: annotation with valid value 'kubernetes'",
			annotations: map[string]string{
				"cilium.giantswarm.io/ipam-mode": "kubernetes",
			},
			err: nil,
		},
		{
			name: "case 5: annotation with invalid value",
			annotations: map[string]string{
				"cilium.giantswarm.io/ipam-mode": "wrong",
			},
			err: notAllowedError,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := &Validator{
				k8sClient: unittest.FakeK8sClient(),
				logger:    microloggertest.New(),
			}
			cluster := unittest.DefaultCluster()
			cluster.Annotations = tc.annotations

			// check if the result is as expected
			err := v.ValidateCiliumIpamMode(cluster)
			if microerror.Cause(err) != tc.err {
				t.Fatalf("Expected error to be %q, got %q", microerror.Cause(tc.err), err)
			}
		})
	}
}

func Test_ValidateCiliumIpamModeUnchanged(t *testing.T) {
	testCases := []struct {
		name           string
		oldAnnotations map[string]string
		oldRelease     string
		newAnnotations map[string]string
		newRelease     string
		err            error
	}{
		{
			name:           "case 0: old release not using cilium, no annotation set",
			oldAnnotations: nil,
			oldRelease:     "18.0.0",
			newAnnotations: nil,
			newRelease:     "18.0.0",
			err:            nil,
		},
		{
			name:           "case 1: old release using cilium, no annotation set",
			oldAnnotations: nil,
			oldRelease:     "19.0.0",
			newAnnotations: nil,
			newRelease:     "19.0.0",
			err:            nil,
		},
		{
			name:           "case 2: old release not using cilium, version unchanged and annotation unchanged",
			oldAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			oldRelease:     "18.0.0",
			newAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			newRelease:     "18.0.0",
			err:            nil,
		},
		{
			name:           "case 3: old release not using cilium, version unchanged and annotation changed",
			oldAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			oldRelease:     "18.0.0",
			newAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "eni"},
			newRelease:     "18.0.0",
			err:            nil,
		},
		{
			name:           "case 4: old release not using cilium, version changed and annotation changed",
			oldAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			oldRelease:     "18.0.0",
			newAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "eni"},
			newRelease:     "19.0.0",
			err:            nil,
		},
		{
			name:           "case 5: old release using cilium, version unchanged and annotation changed",
			oldAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			oldRelease:     "19.0.0",
			newAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "eni"},
			newRelease:     "19.0.0",
			err:            notAllowedError,
		},
		{
			name:           "case 6: old release using cilium, version unchanged and annotation unchanged",
			oldAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			oldRelease:     "19.0.0",
			newAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			newRelease:     "19.0.0",
			err:            nil,
		},
		{
			name:           "case 7: old release using cilium, version unchanged and annotation removed",
			oldAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			oldRelease:     "19.0.0",
			newAnnotations: nil,
			newRelease:     "19.0.0",
			err:            notAllowedError,
		},
		{
			name:           "case 8: old release using cilium, version unchanged and annotation added with non-default value",
			oldAnnotations: nil,
			oldRelease:     "19.0.0",
			newAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "eni"},
			newRelease:     "19.0.0",
			err:            notAllowedError,
		},
		{
			name:           "case 9: old release using cilium, version unchanged and annotation added with default value",
			oldAnnotations: nil,
			oldRelease:     "19.0.0",
			newAnnotations: map[string]string{"cilium.giantswarm.io/ipam-mode": "kubernetes"},
			newRelease:     "19.0.0",
			err:            nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := &Validator{
				k8sClient: unittest.FakeK8sClient(),
				logger:    microloggertest.New(),
			}
			oldCluster := unittest.DefaultCluster()
			oldCluster.Annotations = tc.oldAnnotations
			oldCluster.Labels[label.ReleaseVersion] = tc.oldRelease

			newCluster := unittest.DefaultCluster()
			newCluster.Annotations = tc.newAnnotations
			newCluster.Labels[label.ReleaseVersion] = tc.newRelease

			// check if the result is as expected
			err := v.ValidateCiliumIpamModeUnchanged(oldCluster, newCluster)
			if microerror.Cause(err) != tc.err {
				t.Fatalf("Expected error to be %q, got %q", microerror.Cause(tc.err), err)
			}
		})
	}
}
