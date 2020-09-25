package g8scontrolplane

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

var (
	controlPlaneName      = "gmk24"
	controlPlaneNameSpace = "default"
)

func TestAZG8sControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name                    string
		ctx                     context.Context
		currentAvailabilityZone []string
		dryRun                  bool
		// expectAvailabilityZones needs to be in order
		expectAvailabilityZones []string
		validAvailabilityZones  []string
	}{
		{
			// Update from 1 to 3 Masters with 3 valid AZs
			name:                    "case 0",
			ctx:                     context.Background(),
			currentAvailabilityZone: []string{"eu-central-1a"},
			expectAvailabilityZones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			validAvailabilityZones:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			// Update from 1 to 3 Masters with 2 valid AZs
			name:                    "case 1",
			ctx:                     context.Background(),
			currentAvailabilityZone: []string{"cn-north-1a"},
			expectAvailabilityZones: []string{"cn-north-1a", "cn-north-1b", "cn-north-1a"},
			validAvailabilityZones:  []string{"cn-north-1a", "cn-north-1b"},
		},
		{
			// Update from 1 to 3 Masters with 1 valid AZ
			name:                    "case 2",
			ctx:                     context.Background(),
			currentAvailabilityZone: []string{"cn-south-1a"},
			expectAvailabilityZones: []string{"cn-south-1a", "eu-south-1a", "cn-south-1a"},
			validAvailabilityZones:  []string{"cn-south-1a"},
		},
		{
			name:                    "case 3",
			ctx:                     context.Background(),
			dryRun:                  true,
			currentAvailabilityZone: []string{"cn-south-1a"},
			expectAvailabilityZones: []string{"cn-south-1a"},
			validAvailabilityZones:  []string{"cn-south-1a"},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

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
				validAvailabilityZones: tc.validAvailabilityZones,
				k8sClient:              fakeK8sClient,
				logger:                 newLogger,
			}

			// create AWSControlPlane with the current AZ which belongs to G8sControlPlane
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsControlPlane(tc.currentAvailabilityZone))
			if err != nil {
				t.Fatal(err)
			}

			// run admission request to update AWSControlPlane AZ's
			request, err := g8sControlPlaneUpdateAdmissionRequest(tc.dryRun)
			if err != nil {
				t.Fatal(err)
			}
			_, err = mutate.Mutate(request)
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

func g8sControlPlaneUpdateAdmissionRequest(dryRun bool) (*v1beta1.AdmissionRequest, error) {
	g8scontrolplane, err := getG8sControlPlaneRAWByte(3, "11.5.0")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	g8scontrolplaneOld, err := getG8sControlPlaneRAWByte(1, "11.5.0")
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
		DryRun:    Bool(dryRun),
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    g8scontrolplane,
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    g8scontrolplaneOld,
			Object: nil,
		},
	}
	return req, nil
}

func getG8sControlPlaneRAWByte(replicaNum int, release string) ([]byte, error) {
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
				"release.giantswarm.io/version": release,
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
	byt, err := json.Marshal(g8scontrolPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return byt, nil
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

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}
