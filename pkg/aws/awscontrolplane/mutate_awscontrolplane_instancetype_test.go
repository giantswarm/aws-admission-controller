package awscontrolplane

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-admission-controller/pkg/aws"
	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestInstanceTypeAWSControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name                 string
		ctx                  context.Context
		currentInstanceType  string
		expectedInstanceType string
	}{
		{
			// Don't default the InstanceType if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentInstanceType:  "m4.xlarge",
			expectedInstanceType: "",
		},
		{
			// Default the InstanceType if it is set
			name: "case 1",
			ctx:  context.Background(),

			currentInstanceType:  "",
			expectedInstanceType: aws.DefaultMasterInstanceType,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedInstanceType string

			// Create a new logger that is used by all admitters.
			var newLogger micrologger.Logger
			{
				newLogger, err = micrologger.New(micrologger.Config{})
				if err != nil {
					panic(microerror.JSON(err))
				}
			}
			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				validAvailabilityZones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
				k8sClient:              fakeK8sClient,
				logger:                 newLogger,
			}

			// run admission request to default AWSControlPlane InstanceType
			var patch []mutator.PatchOperation
			request, err := awsControlPlaneAdmissionRequest([]string{"eu-central-1a", "eu-central-1b", "eu-central-1c"}, tc.currentInstanceType, HAReleaseVersion)
			if err != nil {
				t.Fatal(err)
			}
			patch, err = mutate.Mutate(request)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/instanceType" {
					updatedInstanceType = p.Value.(string)
				}
			}
			// check if the instanceType is as expected
			if tc.expectedInstanceType != updatedInstanceType {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedInstanceType, updatedInstanceType)
			}
		})
	}
}
