package aws

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
	"github.com/giantswarm/micrologger/microloggertest"
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

func TestValidateLabelValues(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		newLabels map[string]string
		valid     bool
	}{
		{
			// no label was changed
			name: "case 0",
			ctx:  context.Background(),

			newLabels: unittest.DefaultLabels(),
			valid:     true,
		},
		{
			// version label was changed
			name: "case 1",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                unittest.DefaultClusterID,
				label.ClusterOperatorVersion: unittest.DefaultClusterOperatorVersion,
				label.Release:                "0.0.0",
				label.Organization:           "example-organization",
			},
			valid: true,
		},
		{
			// label was changed
			name: "case 2",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                unittest.DefaultClusterID,
				label.ClusterOperatorVersion: unittest.DefaultClusterOperatorVersion,
				label.Release:                unittest.DefaultReleaseVersion,
				label.Organization:           "new-organization",
			},
			valid: false,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			handle := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			oldObject := unittest.DefaultCluster()
			newObject := unittest.DefaultCluster()
			newObject.SetLabels(tc.newLabels)
			err = ValidateLabelValues(handle, &oldObject, &newObject)
			// check if the result is as expected
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}
