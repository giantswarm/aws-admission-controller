package cluster

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	aws "github.com/giantswarm/aws-admission-controller/v3/pkg/aws/v1alpha3"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v3/pkg/unittest/v1alpha3"
)

func TestMutateOperatorVersion(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentOperator string
		expectedPatch   string
	}{
		{
			// Don't default the Operator Label if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentOperator: unittest.DefaultClusterOperatorVersion,
			expectedPatch:   "",
		},
		{
			// Default the Operator Label if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentOperator: "",
			expectedPatch:   unittest.DefaultClusterOperatorVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedOperator string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}
			// create release
			release := unittest.DefaultRelease()
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &release)
			if err != nil {
				t.Fatal(err)
			}
			// run mutate function to default cluster operator label
			var patch []mutator.PatchOperation
			cluster := unittest.DefaultCluster()
			cluster.SetLabels(map[string]string{label.ClusterOperatorVersion: tc.currentOperator, label.Release: unittest.DefaultReleaseVersion})
			releaseVersion, err := aws.ReleaseVersion(cluster.GetObjectMeta(), patch)
			if err != nil {
				t.Fatal(err)
			}

			patch, err = mutate.MutateOperatorVersion(*cluster, releaseVersion)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", aws.EscapeJSONPatchString(label.ClusterOperatorVersion)) {
					updatedOperator = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedOperator {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedOperator)
			}
		})
	}
}

func TestMutateReleaseUpdate(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		newVersion    string
		oldVersion    string
		expectedPatch string
	}{
		{
			// Don't default the Operator Label if there is no upgrade
			name: "case 0",
			ctx:  context.Background(),

			newVersion:    unittest.DefaultReleaseVersion,
			oldVersion:    unittest.DefaultReleaseVersion,
			expectedPatch: "",
		},
		{
			// Default the Operator Label if there is an upgrade
			name: "case 1",
			ctx:  context.Background(),

			newVersion:    unittest.DefaultReleaseVersion,
			oldVersion:    "99.9.9",
			expectedPatch: unittest.DefaultClusterOperatorVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedOperator string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}
			// create release
			release := unittest.DefaultRelease()
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &release)
			if err != nil {
				t.Fatal(err)
			}

			// create old and new objects
			cluster := unittest.DefaultCluster()
			oldCluster := unittest.DefaultCluster()
			cluster.SetLabels(map[string]string{label.Release: tc.newVersion})
			oldCluster.SetLabels(map[string]string{label.Release: tc.oldVersion})

			// run mutate function to default cluster operator label
			var patch []mutator.PatchOperation
			patch, err = mutate.MutateReleaseUpdate(*cluster, *oldCluster)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", aws.EscapeJSONPatchString(label.ClusterOperatorVersion)) {
					updatedOperator = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedOperator {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedOperator)
			}
		})
	}
}
