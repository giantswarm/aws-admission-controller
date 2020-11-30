package aws

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
)

func Test_MaxBatchSizeIsValid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "case 0: int - simple value",
			input: "5",
			valid: true,
		},
		{
			name:  "case 1: int - big value",
			input: "200",
			valid: true,
		},
		{
			name:  "case 2: int - invalid value - negative number",
			input: "-10",
			valid: false,
		},
		{
			name:  "case 2: int - invalid value - zero",
			input: "0",
			valid: false,
		},
		{
			name:  "case 4: percentage - simple value",
			input: "0.5",
			valid: true,
		},
		{
			name:  "case 5: percentage - rounding",
			input: "0.35",
			valid: true,
		},
		{
			name:  "case 6: percentage - rounding",
			input: "0.32",
			valid: true,
		},
		{
			name:  "case 7: percentage - invalid value - too big",
			input: "1.5",
			valid: false,
		},
		{
			name:  "case 8: percentage - invalid value - negative",
			input: "-0.5",
			valid: false,
		},
		{
			name:  "case 9: invalid value - '50%'",
			input: "50%",
			valid: false,
		},
		{
			name:  "case 10: invalid value - string",
			input: "test",
			valid: false,
		},
		{
			name:  "case 11: invalid value - number and string",
			input: "5erft",
			valid: false,
		},
		{
			name:  "case 12: invalid value - float and string",
			input: "0.5erft",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			isValid := MaxBatchSizeIsValid(tc.input)

			if isValid != tc.valid {
				t.Fatalf("%s - expected '%t' got '%t'\n", tc.name, tc.valid, isValid)
			}
		})
	}
}

func Test_PauseTimeIsValid(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "case 0: simple value",
			value: "PT15M",
			valid: true,
		},
		{
			name:  "case 2: simple value",
			value: "PT10S",
			valid: true,
		},
		{
			name:  "case 3: simple value",
			value: "PT2M10S",
			valid: true,
		},
		{
			name:  "case 4: simple value",
			value: "PT2M10S",
			valid: true,
		},
		{
			name:  "case 5: invalid value value",
			value: "10m",
			valid: false,
		},
		{
			name:  "case 6: invalid value value",
			value: "10s",
			valid: false,
		},
		{
			name:  "case 7: invalid value value",
			value: "10",
			valid: false,
		},
		{
			name:  "case 8: invalid value value",
			value: "1 hour",
			valid: false,
		},
		{
			name:  "case 9: invalid value value",
			value: "random string",
			valid: false,
		},
		{
			name:  "case 10: invalid value value",
			value: "",
			valid: false,
		},
		{
			name:  "case 11: duration too big",
			value: "PT1H2M",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := PauseTimeIsValid(tc.value)

			if result != tc.valid {
				t.Fatalf("%s -  expected '%t' got '%t'\n", tc.name, tc.valid, result)
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
			patch, err = MutateLabelFromCluster(mutate, awscluster.GetObjectMeta(), cluster, label.Release)
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
