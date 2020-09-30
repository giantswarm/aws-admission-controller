package awsmachinedeployment

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-admission-controller/pkg/label"
	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

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
