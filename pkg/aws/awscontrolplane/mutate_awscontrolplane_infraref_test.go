package awscontrolplane

import (
	"context"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestInfraRefAWSControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context

		expectedReferenceName string
	}{
		{
			// Basic test to see whether the infrastructure reference is set
			name: "case 0",
			ctx:  context.Background(),

			expectedReferenceName: controlPlaneName,
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
			mutator := &Mutator{
				validAvailabilityZones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
				k8sClient:              fakeK8sClient,
				logger:                 newLogger,
			}

			// create G8sControlPlane
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, g8sControlPlaneWithoutInfraRef())
			if err != nil {
				t.Fatal(err)
			}
			// run admission request to update the infrastructure reference
			request, err := awsControlPlaneAdmissionRequest([]string{"eu-central-1a"}, "m4.xlarge", HAReleaseVersion)
			if err != nil {
				t.Fatal(err)
			}
			_, err = mutator.Mutate(request)
			if err != nil {
				t.Fatal(err)
			}

			// get G8sControlPlane to verify it has been updated
			updatedG8sControlPlane := &infrastructurev1alpha2.G8sControlPlane{}
			err = fakeK8sClient.CtrlClient().Get(
				tc.ctx,
				types.NamespacedName{
					Name:      controlPlaneName,
					Namespace: controlPlaneNameSpace,
				},
				updatedG8sControlPlane,
			)
			if err != nil {
				t.Fatal(err)
			}

			// check if the InfraRef has been set
			if updatedG8sControlPlane.Spec.InfrastructureRef.Name != tc.expectedReferenceName {
				t.Fatalf("expected %#q to be equal to %#q", updatedG8sControlPlane.Spec.InfrastructureRef.Name, tc.expectedReferenceName)
			}

		})
	}
}

func g8sControlPlaneWithoutInfraRef() *infrastructurev1alpha2.G8sControlPlane {
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
			Replicas:          1,
			InfrastructureRef: v1.ObjectReference{},
		},
	}
	return g8scontrolPlane
}
