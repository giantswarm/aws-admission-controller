package awscontrolplane

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-admission-controller/pkg/label"
	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestAZReplicaMatch(t *testing.T) {
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

			allowed:  false,
			replicas: 3,
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
				logger:    microloggertest.New(),
			}

			g8sControlPlane := unittest.DefaultG8sControlPlane()
			g8sControlPlane.Spec.Replicas = tc.replicas

			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &g8sControlPlane)
			if err != nil {
				t.Fatal(err)
			}

			admissionRequest, err := unittest.DefaultAdmissionRequestAWSControlPlane()
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

func TestAZCount(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed bool
		azs     []string
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed: true,
			azs:     []string{"cn-south-1a"},
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed: true,
			azs:     []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			ctx:  context.Background(),
			name: "case 3",

			allowed: false,
			azs:     []string{"cn-south-1a", "cn-south-1a", "cn-south-1a", "cn-south-1b"},
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

			admissionRequest, err := awsControlPlaneAdmissionRequest(tc.azs, "m4.xlarge", "100.0.0")
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

func TestControlPlaneLabelMatch(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed        bool
		controlPlaneID string
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed:        true,
			controlPlaneID: unittest.DefaultControlPlaneID,
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed:        false,
			controlPlaneID: "notFound",
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

			g8sControlPlane := unittest.DefaultG8sControlPlane()
			g8sControlPlane.SetLabels(map[string]string{label.ControlPlane: tc.controlPlaneID})

			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &g8sControlPlane)
			if err != nil {
				t.Fatal(err)
			}

			admissionRequest, err := unittest.DefaultAdmissionRequestAWSControlPlane()
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
