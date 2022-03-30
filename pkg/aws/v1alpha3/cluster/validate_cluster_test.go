package cluster

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/apiextensions/v6/pkg/label"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/micrologger/microloggertest"
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	unittest "github.com/giantswarm/aws-admission-controller/v3/pkg/unittest/v1alpha3"
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
