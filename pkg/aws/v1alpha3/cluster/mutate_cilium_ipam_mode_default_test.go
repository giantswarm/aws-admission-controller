package cluster

import (
	"context"
	"strconv"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/micrologger/microloggertest"
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"

	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v4/pkg/unittest/v1alpha3"
)

func TestDefaultCiliumIpamMode(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		release          string
		currentValue     string
		expectedIpamMode string
	}{
		{
			name: "case 0: pre-cilium release",
			ctx:  context.Background(),

			release:          "18.0.0",
			expectedIpamMode: "",
		},
		{
			name: "case 1: cilium release, no previous value",
			ctx:  context.Background(),

			release:          "19.0.0",
			currentValue:     "",
			expectedIpamMode: "kubernetes",
		},
		{
			name: "case 2: cilium release, existing value",
			ctx:  context.Background(),

			release:          "19.0.0",
			currentValue:     "eni",
			expectedIpamMode: "",
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient:            fakeK8sClient,
				logger:               microloggertest.New(),
				defaultAWSCNIPodCidr: "10.10.0.0/16",
				defaultCiliumPodCidr: "192.168.0.0/16",
			}
			// create releases
			releases := []releasev1alpha1.Release{
				unittest.NamedRelease("v18.0.0"),
				unittest.NamedRelease("v19.0.0"),
			}
			for _, release := range releases {
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, &release) // nolint:gosec
				if err != nil {
					t.Fatal(err)
				}
			}

			// create old and new objects
			cluster := unittest.DefaultCluster()
			cluster.SetLabels(map[string]string{label.Cluster: cluster.Labels[label.Cluster], label.Release: tc.release})

			if tc.currentValue != "" {
				if cluster.Annotations == nil {
					cluster.Annotations = make(map[string]string, 0)
				}
				cluster.Annotations[annotation.CiliumIpamModeAnnotation] = tc.currentValue
			}

			// run mutate function to default cluster operator label
			var patch []mutator.PatchOperation
			release := semver.MustParse(tc.release)

			patch = mutate.DefaultCiliumIpamMode(*cluster, &release)

			// parse patches
			var ipamMode string
			for _, p := range patch {
				if p.Path == "/metadata/annotations" {
					annotations := p.Value.(map[string]string)

					ipamMode = annotations[annotation.CiliumIpamModeAnnotation]
				}
			}
			// check if the ipam mode is as expected
			if tc.expectedIpamMode != ipamMode {
				t.Fatalf("%s: expected %#q, got %#q", tc.name, tc.expectedIpamMode, ipamMode)
			}
		})
	}
}
