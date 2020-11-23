package awsmachinedeployment

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
)

func TestAWSMachineDeploymentAvailabilityZones(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentAZ     []string
		expectedPatch []string
	}{
		{
			// Don't default the AZ if they are set
			name: "case 0",
			ctx:  context.Background(),

			currentAZ:     unittest.DefaultAvailabilityZones(),
			expectedPatch: nil,
		},
		{
			// Default the AZ they are not set
			name: "case 1",
			ctx:  context.Background(),

			currentAZ:     nil,
			expectedPatch: []string{unittest.DefaultMasterAvailabilityZone},
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedAZs []string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}
			awsControlPlane := unittest.DefaultAWSControlPlane()
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &awsControlPlane)
			if err != nil {
				t.Fatal(err)
			}

			// run mutate function to default AWSMachineDeployment AZs
			var patch []mutator.PatchOperation
			awsmachinedeployment := unittest.DefaultAWSMachineDeployment()
			awsmachinedeployment.Spec.Provider.AvailabilityZones = tc.currentAZ
			patch, err = mutate.MutateAvailabilityZones(awsmachinedeployment)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/provider/availabilityZones" {
					updatedAZs = p.Value.([]string)
				}
			}

			// check if the AZs are patched as expected
			if len(tc.expectedPatch) != len(updatedAZs) {
				t.Fatalf("expected %v to not to differ from %v", len(tc.expectedPatch), len(updatedAZs))
			}
			for i, p := range updatedAZs {
				if tc.expectedPatch[i] != p {
					t.Fatalf("expected %v to not to differ from %v", tc.expectedPatch[i], p)
				}
			}
		})
	}
}

func TestMachineDeploymentLabelMatch(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		allowed             bool
		machineDeploymentID string
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			allowed:             true,
			machineDeploymentID: unittest.DefaultMachineDeploymentID,
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			allowed:             false,
			machineDeploymentID: "notFound",
		},
		// empty case
		{
			ctx:  context.Background(),
			name: "case 2",

			allowed: false,
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

			machineDeployment := unittest.DefaultMachineDeployment()
			machineDeployment.SetLabels(map[string]string{label.MachineDeployment: tc.machineDeploymentID})

			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &machineDeployment)
			if err != nil {
				t.Fatal(err)
			}

			admissionRequest, err := unittest.DefaultAdmissionRequestAWSMachineDeployment()
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
