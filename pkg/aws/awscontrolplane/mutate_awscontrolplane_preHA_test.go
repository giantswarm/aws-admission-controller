package awscontrolplane

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestPreHAAWSControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name                    string
		ctx                     context.Context
		currentAvailabilityZone []string
		clusterAvailabilityZone string
		expectAvailabilityZones []string
		currentInstanceType     string
		clusterInstanceType     string
		expectInstanceType      string
	}{
		{
			// Default AZ if it is not set
			name: "case 0",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			clusterAvailabilityZone: "cn-south-1a",
			expectAvailabilityZones: []string{"cn-south-1a"},
		},
		{
			// Do not default AZ if it is set
			name: "case 1",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"cn-south-1a"},
			clusterAvailabilityZone: "cn-south-1a",
			expectAvailabilityZones: nil,
		},
		{
			// Default InstanceType if it is not set
			name: "case 2",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"cn-south-1a"},
			currentInstanceType:     "",
			clusterInstanceType:     "m4.xlarge",
			expectInstanceType:      "m4.xlarge",
		},
		{
			// Do not default InstanceType if it is set
			name: "case 3",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"cn-south-1a"},
			currentInstanceType:     "m4.xlarge",
			expectInstanceType:      "",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedAZs []string
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
				validAvailabilityZones: []string{"cn-south-1a"},
				k8sClient:              fakeK8sClient,
				logger:                 newLogger,
			}

			// create AWSCluster
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsCluster(tc.clusterInstanceType, tc.clusterAvailabilityZone))
			if err != nil {
				t.Fatal(err)
			}
			// run admission request to default AWSControlPlane
			var patch []mutator.PatchOperation
			request, err := awsControlPlaneAdmissionRequest(tc.currentAvailabilityZone, tc.currentInstanceType, preHAReleaseVersion)
			if err != nil {
				t.Fatal(err)
			}
			patch, err = mutate.Mutate(request)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/availabilityZones" {
					updatedAZs = p.Value.([]string)
				} else if p.Path == "/spec/instanceType" {
					updatedInstanceType = p.Value.(string)
				}
			}
			// check if the AZ is as expected
			if len(tc.expectAvailabilityZones) != len(updatedAZs) {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectAvailabilityZones, updatedAZs)
			} else if len(tc.expectAvailabilityZones) > 0 {
				if tc.expectAvailabilityZones[0] != updatedAZs[0] {
					t.Fatalf("expected %#q to be equal to %#q", tc.expectAvailabilityZones, updatedAZs)
				}
			}
			// check if the instanceType is as expected
			if tc.expectInstanceType != updatedInstanceType {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectInstanceType, updatedInstanceType)
			}
		})
	}
}

func awsCluster(instanceType string, availabilityZone string) *infrastructurev1alpha2.AWSCluster {
	awsCluster := &infrastructurev1alpha2.AWSCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSCluster",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/control-plane":   controlPlaneName,
				"giantswarm.io/cluster":         clusterName,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": preHAReleaseVersion,
			},
		},
		Spec: infrastructurev1alpha2.AWSClusterSpec{
			Provider: infrastructurev1alpha2.AWSClusterSpecProvider{
				Master: infrastructurev1alpha2.AWSClusterSpecProviderMaster{
					InstanceType:     instanceType,
					AvailabilityZone: availabilityZone,
				},
			},
		},
	}
	return awsCluster
}
