package machinedeployment

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	unittest "github.com/giantswarm/aws-admission-controller/v3/pkg/unittest/v1alpha3"
)

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
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, cluster)
			if err != nil {
				t.Fatal(err)
			}

			// try to create the machinedeployment
			object := unittest.DefaultMachineDeployment()
			err = validate.ValidateCluster(*object)
			if tc.allowed && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.allowed && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func TestValidateClusterNamespace(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		clusterNamespace  string
		nodePoolNamespace string
		allowed           bool
	}{
		{
			ctx:  context.Background(),
			name: "case 0",

			clusterNamespace:  "default",
			nodePoolNamespace: "default",
			allowed:           true,
		},
		{
			ctx:  context.Background(),
			name: "case 1",

			clusterNamespace:  "org-test",
			nodePoolNamespace: "org-test",
			allowed:           true,
		},
		{
			ctx:  context.Background(),
			name: "case 2",

			clusterNamespace:  "default",
			nodePoolNamespace: "org-test",
			allowed:           false,
		},
		{
			ctx:  context.Background(),
			name: "case 3",

			clusterNamespace:  "org-test",
			nodePoolNamespace: "default",
			allowed:           false,
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
			cluster.SetNamespace(tc.clusterNamespace)
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, cluster)
			if err != nil {
				t.Fatal(err)
			}

			// try to create the awsmachinedeployment
			object := unittest.DefaultMachineDeployment()
			object.SetNamespace(tc.nodePoolNamespace)
			err = validate.ValidateCluster(*object)
			if tc.allowed && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.allowed && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}
