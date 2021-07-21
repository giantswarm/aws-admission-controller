package v1alpha2

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v2/pkg/unittest/v1alpha2"
)

func TestLabel(t *testing.T) {
	testCases := []struct {
		name string

		label         string
		defaultValue  string
		currentValue  string
		expectedPatch string
	}{
		{
			// Don't default the Label if it is set
			name: "case 0",

			label:         label.ControlPlane,
			defaultValue:  unittest.DefaultControlPlaneID,
			currentValue:  "abcd",
			expectedPatch: "",
		},
		{
			// Don't default the Label if it is set
			name: "case 1",

			label:         label.ControlPlane,
			defaultValue:  unittest.DefaultControlPlaneID,
			currentValue:  "",
			expectedPatch: unittest.DefaultControlPlaneID,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedValue string
			var patch []mutator.PatchOperation

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			// run mutate function to default AWSControlplane operator label

			awscontrolplane := unittest.DefaultAWSControlPlane()
			awscontrolplane.SetLabels(map[string]string{tc.label: tc.currentValue})
			patch, err = MutateLabel(mutate, &awscontrolplane, tc.label, tc.defaultValue)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(tc.label)) {
					updatedValue = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedValue {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedValue)
			}
		})
	}
}

func TestLabelFromCluster(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentRelease string
		expectedPatch  string
	}{
		{
			// Don't default the Release Label if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentRelease: unittest.DefaultReleaseVersion,
			expectedPatch:  "",
		},
		{
			// Default the Release Label if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentRelease: "",
			expectedPatch:  unittest.DefaultReleaseVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedRelease string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			// run mutate function to default AWSCluster ReleaseVersion label
			var patch []mutator.PatchOperation
			cluster := unittest.DefaultCluster()
			awscluster := unittest.DefaultAWSCluster()
			awscluster.SetLabels(map[string]string{label.Release: tc.currentRelease, label.Cluster: unittest.DefaultClusterID})
			patch, err = MutateLabelFromCluster(mutate, awscluster.GetObjectMeta(), *cluster, label.Release)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.Release)) {
					updatedRelease = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedRelease {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedRelease)
			}
		})
	}
}

func TestLabelFromAWSCluster(t *testing.T) {
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

			currentOperator: unittest.DefaultAWSOperatorVersion,
			expectedPatch:   "",
		},
		{
			// Default the Operator Label if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentOperator: "",
			expectedPatch:   unittest.DefaultAWSOperatorVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedOperator string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			// run mutate function to default AWSControlplane operator label
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			awscontrolplane := unittest.DefaultAWSControlPlane()
			awscontrolplane.SetLabels(map[string]string{label.AWSOperatorVersion: tc.currentOperator, label.Cluster: unittest.DefaultClusterID})
			patch, err = MutateLabelFromAWSCluster(mutate, awscontrolplane.GetObjectMeta(), awscluster, label.AWSOperatorVersion)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.AWSOperatorVersion)) {
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

func TestLabelFromRelease(t *testing.T) {
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

			currentOperator: unittest.DefaultAWSOperatorVersion,
			expectedPatch:   "",
		},
		{
			// Default the Operator Label if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentOperator: "",
			expectedPatch:   unittest.DefaultAWSOperatorVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedOperator string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			// run mutate function to default AWSControlplane operator label
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			release := unittest.DefaultRelease()
			awscluster.SetLabels(map[string]string{label.AWSOperatorVersion: tc.currentOperator, label.Release: unittest.DefaultReleaseVersion})
			patch, err = MutateLabelFromRelease(mutate, awscluster.GetObjectMeta(), release, label.AWSOperatorVersion, "aws-operator")
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.AWSOperatorVersion)) {
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
