package g8scontrolplane

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestReplicaAZMatch(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed bool
		AZs     int
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed: true,
			AZs:     1,
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed: false,
			AZs:     3,
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
			validate := &Validator{
				k8sClient: fakeK8sClient,
				logger:    newLogger,
			}

			awsControlPlane := unittest.DefaultAWSControlPlane()
			if tc.AZs == 3 {
				awsControlPlane.Spec.AvailabilityZones = []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"}
			}

			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &awsControlPlane)
			if err != nil {
				t.Fatal(err)
			}

			admissionRequest, err := unittest.DefaultAdmissionRequestG8sControlPlane()
			if err != nil {
				t.Fatal(err)
			}

			allowed, _ := validate.Validate(&admissionRequest)
			if allowed != tc.allowed {
				t.Fatalf("expected %v to not to differ from %v", allowed, tc.allowed)
			}
		})
	}
}

func TestReplicaCount(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed  bool
		replicas int
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed:  true,
			replicas: 1,
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed:  true,
			replicas: 3,
		},
		{
			ctx:  context.Background(),
			name: "case 3",

			allowed:  false,
			replicas: 4,
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
			validate := &Validator{
				k8sClient: fakeK8sClient,
				logger:    newLogger,
			}

			admissionRequest, err := g8sControlPlaneCreateAdmissionRequest(tc.replicas, "100.0.0")
			if err != nil {
				t.Fatal(err)
			}

			allowed, _ := validate.Validate(admissionRequest)
			if allowed != tc.allowed {
				t.Fatalf("expected %v to not to differ from %v", allowed, tc.allowed)
			}
		})
	}
}
