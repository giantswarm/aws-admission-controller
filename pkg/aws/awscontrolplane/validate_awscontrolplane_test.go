package awscontrolplane

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
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

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				validAvailabilityZones: unittest.DefaultAvailabilityZones(),
				validInstanceTypes:     unittest.DefaultInstanceTypes(),
				k8sClient:              fakeK8sClient,
				logger:                 microloggertest.New(),
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
			azs:     []string{"eu-central-1a"},
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
			azs:     []string{"eu-central-1a", "eu-central-1a", "eu-central-1a", "eu-central-1c"},
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				validAvailabilityZones: unittest.DefaultAvailabilityZones(),
				validInstanceTypes:     unittest.DefaultInstanceTypes(),
				k8sClient:              fakeK8sClient,
				logger:                 microloggertest.New(),
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
func TestAZOrder(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed bool
		oldAZs  []string
		newAZs  []string
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed: true,
			oldAZs:  []string{"eu-central-1a"},
			newAZs:  []string{"eu-central-1a"},
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed: true,
			oldAZs:  []string{"eu-central-1a"},
			newAZs:  []string{"eu-central-1b"},
		},
		{
			ctx:  context.Background(),
			name: "case 2",

			allowed: true,
			oldAZs:  []string{"eu-central-1a"},
			newAZs:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			ctx:  context.Background(),
			name: "case 3",

			allowed: false,
			oldAZs:  []string{"eu-central-1a"},
			newAZs:  []string{"eu-central-1b", "eu-central-1a", "eu-central-1c"},
		},
		{
			ctx:  context.Background(),
			name: "case 4",

			allowed: false,
			oldAZs:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			newAZs:  []string{"eu-central-1c", "eu-central-1b", "eu-central-1a"},
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				validAvailabilityZones: unittest.DefaultAvailabilityZones(),
				validInstanceTypes:     unittest.DefaultInstanceTypes(),
				k8sClient:              fakeK8sClient,
				logger:                 microloggertest.New(),
			}

			admissionRequest, err := unittest.CustomAdmissionRequestAWSControlPlaneUpdate(tc.oldAZs, tc.newAZs)
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
func TestAZUnique(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed   bool
		chosenAZs []string
		validAZs  []string
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed:   true,
			chosenAZs: []string{"eu-central-1a"},
			validAZs:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed:   true,
			chosenAZs: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			validAZs:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
		{
			ctx:  context.Background(),
			name: "case 2",

			allowed:   true,
			chosenAZs: []string{"eu-central-1a", "eu-central-1b", "eu-central-1b"},
			validAZs:  []string{"eu-central-1a", "eu-central-1b"},
		},
		{
			ctx:  context.Background(),
			name: "case 3",

			allowed:   true,
			chosenAZs: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
			validAZs:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c", "eu-central-1d"},
		},
		{
			ctx:  context.Background(),
			name: "case 4",

			allowed:   false,
			chosenAZs: []string{"eu-central-1a", "eu-central-1b", "eu-central-1b"},
			validAZs:  []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				validAvailabilityZones: tc.validAZs,
				validInstanceTypes:     unittest.DefaultInstanceTypes(),
				k8sClient:              fakeK8sClient,
				logger:                 microloggertest.New(),
			}

			admissionRequest, err := unittest.CustomAdmissionRequestAWSControlPlane(tc.chosenAZs)
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

func TestAZValid(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed  bool
		validAZs []string
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed:  false,
			validAZs: []string{"cn-south-1a", "cn-south-1b"},
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed:  true,
			validAZs: unittest.DefaultAvailabilityZones(),
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				validAvailabilityZones: tc.validAZs,
				validInstanceTypes:     unittest.DefaultInstanceTypes(),
				k8sClient:              fakeK8sClient,
				logger:                 microloggertest.New(),
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
func TestInstanceTypeValid(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed            bool
		validInstanceTypes []string
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed:            false,
			validInstanceTypes: []string{"c5.xlarge", "c5.2xlarge", "m4.xlarge"},
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed:            true,
			validInstanceTypes: unittest.DefaultInstanceTypes(),
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				validInstanceTypes:     tc.validInstanceTypes,
				validAvailabilityZones: unittest.DefaultAvailabilityZones(),
				k8sClient:              fakeK8sClient,
				logger:                 microloggertest.New(),
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
