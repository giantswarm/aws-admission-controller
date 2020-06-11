package g8scontrolplane

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/g8s-admission-controller/pkg/unittest"
)

var (
	controlPlaneName      = "gmk24"
	controlPlaneNameSpace = "default"
)

func TestG8sControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name                    string
		ctx                     context.Context
		currentAvailabilityZone []string
		// expectAvailabilityZones needs to be in order
		expectAvailabilityZones []string
		validAvailabilityZones  []string
	}{
		{
			name: "case 0",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"eu-central-1a"},
			expectAvailabilityZones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			validAvailabilityZones:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			name: "case 1",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"cn-north-1a"},
			expectAvailabilityZones: []string{"cn-north-1a", "cn-north-1b", "cn-north-1a"},
			validAvailabilityZones:  []string{"cn-north-1a", "eu-central-1b"},
		},
		{
			name: "case 2",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"cn-south-1a"},
			expectAvailabilityZones: []string{"cn-south-1a", "eu-south-1a", "cn-south-1a"},
			validAvailabilityZones:  []string{"cn-south-1a"},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			admit := &Admitter{
				validAvailabilityZones: tc.validAvailabilityZones,
				k8sClient:              fakeK8sClient,
			}

			// create AWSControlPlane with the current AZ which belongs to G8sControlPlane
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsControlPlane(tc.currentAvailabilityZone))
			if err != nil {
				t.Fatal(err)
			}

			// run admission request to update AWSControlPlane AZ's
			_, err = admit.Admit(g8sControlPlaneAdmissionRequest())
			if err != nil {
				t.Fatal(err)
			}

			// get AWSControlPlane to verify it has been updated
			updatedAWSControlPlane := &infrastructurev1alpha2.AWSControlPlane{}
			err = fakeK8sClient.CtrlClient().Get(
				tc.ctx,
				types.NamespacedName{
					Name:      controlPlaneName,
					Namespace: controlPlaneNameSpace,
				},
				updatedAWSControlPlane,
			)
			if err != nil {
				t.Fatal(err)
			}

			// sorting again due to shuffling the AZ's
			sort.Strings(updatedAWSControlPlane.Spec.AvailabilityZones)
			updatedAZs := updatedAWSControlPlane.Spec.AvailabilityZones

			// check if the amount of AZ's is correct
			if len(tc.expectAvailabilityZones) != len(updatedAZs) {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectAvailabilityZones, updatedAZs)
			}

			// check if updated AZ's are in expected AZ's
			// if only two valid AZ's, ignore last AZ because it's randomly picked
			for i, az := range updatedAZs {
				if len(tc.validAvailabilityZones) == 2 && i == 2 {
					return
				}
				if !updatedAZinExpectedAZs(az, tc.expectAvailabilityZones) {
					t.Fatalf("expected AZ %s is missing in updated AZ list %v", az, updatedAZs)
				}
			}
		})
	}
}

func g8sControlPlaneAdmissionRequest() *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "G8sControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "g8scontrolplanes",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    getG8sControlPlaneRAWByte(3),
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    getG8sControlPlaneRAWByte(1),
			Object: nil,
		},
	}
	return req
}

func getG8sControlPlaneRAWByte(replicaNum int) []byte {
	g8scontrolPlane := infrastructurev1alpha2.G8sControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "G8sControlPlane",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/control-plane":   controlPlaneName,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": "11.5.0",
			},
		},
		Spec: infrastructurev1alpha2.G8sControlPlaneSpec{
			Replicas: replicaNum,
			InfrastructureRef: v1.ObjectReference{
				Kind:       "AWSControlPlane",
				Namespace:  "default",
				Name:       "gmk24",
				APIVersion: "infrastructure.giantswarm.io/v1alpha2",
			},
		},
	}
	byt, _ := json.Marshal(g8scontrolPlane)
	return byt
}

func awsControlPlane(currentAvailabilityZone []string) *infrastructurev1alpha2.AWSControlPlane {
	awsControlPlane := &infrastructurev1alpha2.AWSControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSControlPlane",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/control-plane":   controlPlaneName,
				"release.giantswarm.io/version": "11.5.0",
			},
		},
		Spec: infrastructurev1alpha2.AWSControlPlaneSpec{
			AvailabilityZones: currentAvailabilityZone,
			InstanceType:      "m4.xlarge",
		},
	}
	return awsControlPlane
}

func updatedAZinExpectedAZs(az string, expectedAZs []string) bool {
	for _, expectAZ := range expectedAZs {
		if expectAZ == az {
			return true
		}
	}
	return false
}
