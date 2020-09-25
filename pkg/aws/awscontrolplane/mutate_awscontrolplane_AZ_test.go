package awscontrolplane

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/ruleengine"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

var (
	controlPlaneName      = "gmk24"
	controlPlaneNameSpace = "default"
	clusterName           = "gmk24"
	preHAReleaseVersion   = "11.3.0"
	HAReleaseVersion      = "11.5.0"
)

func TestAZAWSControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
		// if the Replicas are nil, it means no g8sControlPlane exists
		g8sControlplaneReplicas *int
		currentAvailabilityZone []string
		// expectAvailabilityZones needs to be in order
		expectAvailabilityZones []string
		validAvailabilityZones  []string
	}{
		{
			// Defaulting for 3 valid AZs
			name: "case 0",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			g8sControlplaneReplicas: nil,
			expectAvailabilityZones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			validAvailabilityZones:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			// Defaulting for 2 valid AZs
			name: "case 1",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			g8sControlplaneReplicas: nil,
			expectAvailabilityZones: []string{"cn-north-1a", "cn-north-1b", "cn-north-1a"},
			validAvailabilityZones:  []string{"cn-north-1a", "cn-north-1b"},
		},
		{
			// Defaulting for 1 valid AZ
			name: "case 2",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			g8sControlplaneReplicas: nil,
			expectAvailabilityZones: []string{"cn-south-1a", "eu-south-1a", "cn-south-1a"},
			validAvailabilityZones:  []string{"cn-south-1a"},
		},
		{
			// Defaulting for 3 g8scontrolplane replicas
			name: "case 3",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			g8sControlplaneReplicas: ruleengine.ToIntPtr(3),
			expectAvailabilityZones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			validAvailabilityZones:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			// Defaulting for 1 g8scontrolplane replica
			name: "case 4",
			ctx:  context.Background(),

			currentAvailabilityZone: nil,
			g8sControlplaneReplicas: ruleengine.ToIntPtr(1),
			expectAvailabilityZones: []string{"cn-south-1a"},
			validAvailabilityZones:  []string{"cn-south-1a"},
		},
		{
			// Here we check if there is no defaulting when AZs are != nil. Note that the expected AZs being
			// nil here means that the current AZs stay unchanged. (The patch is nil)
			name: "case 5",
			ctx:  context.Background(),

			currentAvailabilityZone: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			g8sControlplaneReplicas: nil,
			expectAvailabilityZones: nil,
			validAvailabilityZones:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedAZs []string

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

			// create G8sControlPlane if needed
			if tc.g8sControlplaneReplicas != nil {
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, g8sControlPlane(*tc.g8sControlplaneReplicas))
				if err != nil {
					t.Fatal(err)
				}
			}
			// run admission request to default AWSControlPlane AZ's
			var patch []mutator.PatchOperation
			request, err := awsControlPlaneAdmissionRequest(tc.currentAvailabilityZone, "m4.xlarge", HAReleaseVersion)
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
				}
			}

			// check if the amount of AZ's is correct
			if len(tc.expectAvailabilityZones) != len(updatedAZs) {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectAvailabilityZones, updatedAZs)
			}

			// check if defaulted AZ's are in expected AZ's
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

func awsControlPlaneAdmissionRequest(AZs []string, instanceType string, release string) (*v1beta1.AdmissionRequest, error) {
	awscontrolplane, err := getAWSControlPlaneRAWByte(AZs, instanceType, release)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	req := &v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AWSControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "awscontrolplanes",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    awscontrolplane,
			Object: nil,
		},
	}
	return req, nil
}

func g8sControlPlane(replicaNum int) *infrastructurev1alpha2.G8sControlPlane {
	g8scontrolPlane := &infrastructurev1alpha2.G8sControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "G8sControlPlane",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/cluster":         clusterName,
				"giantswarm.io/control-plane":   controlPlaneName,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": HAReleaseVersion,
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
	return g8scontrolPlane
}

func getAWSControlPlaneRAWByte(currentAvailabilityZone []string, currentInstanceType string, release string) ([]byte, error) {
	awsControlPlane := infrastructurev1alpha2.AWSControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSControlPlane",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/cluster":         clusterName,
				"giantswarm.io/control-plane":   controlPlaneName,
				"release.giantswarm.io/version": release,
			},
		},
		Spec: infrastructurev1alpha2.AWSControlPlaneSpec{
			AvailabilityZones: currentAvailabilityZone,
			InstanceType:      currentInstanceType,
		},
	}
	byt, err := json.Marshal(awsControlPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return byt, nil
}

func updatedAZinExpectedAZs(az string, expectedAZs []string) bool {
	for _, expectAZ := range expectedAZs {
		if expectAZ == az {
			return true
		}
	}
	return false
}
