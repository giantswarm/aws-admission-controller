package g8scontrolplane

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/api/core/v1"

	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v3/pkg/unittest/v1alpha3"
)

// TestMutateInfraRefUpdate tests updating v1alpha2 to v1alpha3
func TestMutateInfraRefUpdate(t *testing.T) {
	testCases := []struct {
		ctx            context.Context
		name           string
		releaseVersion string

		apiVersion string
		mutate     bool
	}{
		{
			// g8scontrolplane infraref with release version >= 16.0.0 gets patched with v1alpha3
			ctx:            context.Background(),
			name:           "case 0",
			releaseVersion: "100.0.0",

			apiVersion: "infrastructure.giantswarm.io/v1alpha3",
			mutate:     true,
		},
		{
			// g8scontrolplane infraref with release version < 16.0.0 gets patched with v1alpha2
			ctx:            context.Background(),
			name:           "case 1",
			releaseVersion: "12.0.0",

			apiVersion: "infrastructure.giantswarm.io/v1alpha2",
			mutate:     false,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var apiVersion string
			var patch []mutator.PatchOperation
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			// try to create the g8sControlplane
			g8sControlPlane := unittest.DefaultG8sControlPlane()
			g8sControlPlane.Labels[label.Release] = tc.releaseVersion
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &g8sControlPlane)
			if err != nil {
				t.Fatal(err)
			}

			patch, err = mutate.MutateInfraRef(g8sControlPlane)
			if err != nil {
				t.Fatal(err)
			}

			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/infrastructureRef" {
					apiVersion = p.Value.(*v1.ObjectReference).APIVersion
				}
			}
			// check if the reference patch is as expected
			if tc.apiVersion != apiVersion {
				t.Fatalf("expected %#q to be equal to %#q", tc.apiVersion, apiVersion)

			}
		})
	}
}
