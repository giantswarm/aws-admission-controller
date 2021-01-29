package awsmachinedeployment

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
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

			// try to create the awsmachinedeployment
			object := unittest.DefaultAWSMachineDeployment()
			err = validate.MachineDeploymentLabelMatch(object)
			if tc.allowed && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.allowed && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func TestValidateCluster(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		deleted bool
		allowed bool
	}{
		{
			// cluster is not deleted
			ctx:  context.Background(),
			name: "case 0",

			deleted: false,
			allowed: true,
		},
		{
			// cluster is deleted
			ctx:  context.Background(),
			name: "case 1",

			deleted: true,
			allowed: false,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			// create the cluster
			cluster := unittest.DefaultCluster()
			if tc.deleted {
				cluster.SetDeletionTimestamp(&v1.Time{time.Now()})
			}
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &cluster)
			if err != nil {
				t.Fatal(err)
			}

			// try to create the awsmachinedeployment
			object := unittest.DefaultAWSMachineDeployment()
			err = validate.ValidateCluster(object)
			if tc.allowed && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.allowed && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}
