package v1alpha2

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
)

func TestCAPIReleaseLabel(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentRelease string
		expectedResult bool
	}{
		{
			// Release label is not set
			name: "case 0",
			ctx:  context.Background(),

			currentRelease: "",
			expectedResult: false,
		},
		{
			// CAPI Release label is set
			name: "case 1",
			ctx:  context.Background(),

			currentRelease: "20.0.0-v1alpha3",
			expectedResult: true,
		},
		{
			// CAPI Release label is set
			name: "case 2",
			ctx:  context.Background(),

			currentRelease: "20.0.0",
			expectedResult: true,
		},
		{
			// GS Release label is set
			name: "case 3",
			ctx:  context.Background(),

			currentRelease: "14.0.0",
			expectedResult: false,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cluster := unittest.DefaultCluster()
			cluster.SetLabels(map[string]string{label.Release: tc.currentRelease})
			capi, err := IsCAPIRelease(cluster)
			if err != nil {
				t.Fatal(err)
			}
			// check if the result label is as expected
			if tc.expectedResult != capi {
				t.Fatalf("expected %v to be equal to %v", tc.expectedResult, capi)
			}
		})
	}
}
