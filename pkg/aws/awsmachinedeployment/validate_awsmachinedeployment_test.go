package awsmachinedeployment

import (
	"context"
	"strconv"
	"testing"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
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
				cluster.SetDeletionTimestamp(&v1.Time{Time: time.Now()})
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

func TestValidateMachineDeploymentScaling(t *testing.T) {
	testCases := []struct {
		scaling infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling
		matcher func(error) bool
	}{
		{
			// case 0
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 0,
				Max: 2,
			},
			matcher: IsNotAllowed,
		},
		{
			// case 1
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 4,
				Max: 0,
			},
			matcher: IsNotAllowed,
		},
		{
			// case 2
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 0,
				Max: 0,
			},
			matcher: IsNotAllowed,
		},
		{
			// case 3
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 4,
				Max: 2,
			},
			matcher: IsNotAllowed,
		},
		{
			// case 4
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 1,
				Max: 1,
			},
			matcher: nil,
		},
		{
			// case 5
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 4,
				Max: 4,
			},
			matcher: nil,
		},
		{
			// case 6
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 4,
				Max: 6,
			},
			matcher: nil,
		},
		{
			// case 7
			scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
				Min: 1,
				Max: 10,
			},
			matcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := &Validator{
				k8sClient: unittest.FakeK8sClient(),
				logger:    microloggertest.New(),
			}

			md := unittest.DefaultAWSMachineDeployment()

			md.Spec.NodePool.Scaling = tc.scaling

			err := v.MachineDeploymentScaling(md)
			switch {
			case err == nil && tc.matcher == nil:
				// correct; carry on
			case err != nil && tc.matcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.matcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.matcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}
		})
	}
}
