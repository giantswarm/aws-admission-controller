package g8scontrolplane

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestReplicasG8sControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name                    string
		ctx                     context.Context
		currentAvailabilityZone []string
		currentReplicas         int
		expectReplicas          int
		preHArelease            bool
	}{
		{
			// Default replicas for 1 awscontrolplane AZ
			name: "case 0",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"eu-central-1a"},
			currentReplicas:         0,
			expectReplicas:          1,
		},
		{
			// Default replicas for 3 awscontrolplane AZs
			name: "case 1",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			currentReplicas:         0,
			expectReplicas:          3,
		},
		{
			// Default replicas without awscontrolplane in case it's an HA release
			name: "case 2",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			currentReplicas:         0,
			expectReplicas:          3,
		},
		{
			// Don't default replicas if they are set to 1
			name: "case 3",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			currentReplicas:         1,
			expectReplicas:          0,
		},
		{
			// Don't default replicas if they are set to 3
			name: "case 4",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			currentReplicas:         3,
			expectReplicas:          0,
		},
		{
			// Default replicas without awscontrolplane in case it's not an HA release
			name: "case 5",
			ctx:  context.Background(),

			preHArelease:            true,
			currentAvailabilityZone: nil,
			currentReplicas:         0,
			expectReplicas:          1,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedReplicas int
			var release string

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
			if tc.preHArelease {
				release = "11.3.0"
			} else {
				release = "11.5.0"
			}
			// create AWSControlPlane with the current AZ which belongs to G8sControlPlane if needed
			if tc.currentAvailabilityZone != nil {
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsControlPlane(tc.currentAvailabilityZone))
				if err != nil {
					t.Fatal(err)
				}
			}

			// run admission request to default replicas
			var patch []mutator.PatchOperation
			request, err := g8sControlPlaneCreateAdmissionRequest(tc.currentReplicas, release)
			if err != nil {
				t.Fatal(err)
			}
			patch, err = mutate.Mutate(request)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/replicas" {
					updatedReplicas = p.Value.(int)
				}
			}
			// check if the values of Replicas is as expected
			if tc.expectReplicas != updatedReplicas {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectReplicas, updatedReplicas)
			}
		})
	}
}

func g8sControlPlaneCreateAdmissionRequest(replicas int, release string) (*v1beta1.AdmissionRequest, error) {
	g8scontrolplane, err := getG8sControlPlaneRAWByte(replicas, release)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	req := &v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "G8sControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "g8scontrolplanes",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    g8scontrolplane,
			Object: nil,
		},
	}
	return req, nil
}
