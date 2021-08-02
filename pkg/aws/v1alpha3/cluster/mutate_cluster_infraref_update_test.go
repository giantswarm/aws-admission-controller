package cluster

import (
	"context"
	"strconv"
	"testing"

	"github.com/blang/semver"
	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/api/core/v1"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v2/pkg/unittest/v1alpha3"
)

// TestMutateInfraRefUpdate tests updating v1alpha2 to v1alpha3
func TestMutateInfraRefUpdate(t *testing.T) {
	testCases := []struct {
		ctx            context.Context
		name           string
		releaseVersion *semver.Version

		apiVersion string
		mutate     bool
	}{
		{
			// cluster infraref with release version >= 16.0.0 gets updated
			ctx:  context.Background(),
			name: "case 0",
			releaseVersion: &semver.Version{
				Major: uint64(100),
				Minor: uint64(0),
				Patch: uint64(0),
			},

			apiVersion: "infrastructure.giantswarm.io/v1alpha3",
			mutate:     true,
		},
		{
			// cluster infraref with release version < 16.0.0 does not get patched
			ctx:  context.Background(),
			name: "case 1",
			releaseVersion: &semver.Version{
				Major: uint64(12),
				Minor: uint64(0),
				Patch: uint64(0),
			},

			apiVersion: "",
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

			// try to create the cluster
			cluster := unittest.DefaultCluster()
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, cluster)
			if err != nil {
				t.Fatal(err)
			}

			patch, err = mutate.MutateInfraRef(*cluster, tc.releaseVersion)
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
