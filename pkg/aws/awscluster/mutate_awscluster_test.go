package awscluster

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
)

func TestAWSClusterPodCIDR(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentPodCIDR  string
		expectedPodCIDR string
	}{
		{
			// Don't default the Pod CIDR if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentPodCIDR:  unittest.DefaultPodCIDR,
			expectedPodCIDR: "",
		},
		{
			// Default the Pod CIDR if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentPodCIDR:  "",
			expectedPodCIDR: unittest.DefaultPodCIDR,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedCIDR map[string]string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				podCIDRBlock: unittest.DefaultPodCIDR,
				k8sClient:    fakeK8sClient,
				logger:       microloggertest.New(),
			}

			// run admission request to default AWSCluster Pod CIDR
			var patch []mutator.PatchOperation
			request, err := unittest.CustomAdmissionRequestAWSCluster(tc.currentPodCIDR)
			if err != nil {
				t.Fatal(err)
			}
			patch, err = mutate.Mutate(&request)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/provider/pods" {
					updatedCIDR = p.Value.(map[string]string)
				}
			}
			// check if the pod CIDR is as expected
			if tc.expectedPodCIDR != updatedCIDR["cidrBlock"] {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPodCIDR, updatedCIDR)
			}
		})
	}
}
