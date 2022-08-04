package cluster

import (
	"context"
	"strconv"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/micrologger/microloggertest"
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"

	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v4/pkg/unittest/v1alpha3"
)

func TestCiliumCIDRDefaulting(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		newVersion    string
		oldVersion    string
		customPodCidr bool
		networkPool   bool
		expectedCidr  string
	}{
		{
			name: "case 0: no upgrade",
			ctx:  context.Background(),

			newVersion:   "17.0.0",
			oldVersion:   "17.0.0",
			expectedCidr: "",
		},
		{
			name: "case 1: upgrade to v18",
			ctx:  context.Background(),

			newVersion:   "16.0.0",
			oldVersion:   "17.0.0",
			expectedCidr: "",
		},
		{
			name: "case 2: minor upgrade within v18",
			ctx:  context.Background(),

			newVersion:   "18.0.0",
			oldVersion:   "18.0.1",
			expectedCidr: "",
		},
		{
			name: "case 3: upgrade from v17 to v18",
			ctx:  context.Background(),

			newVersion:   "17.0.0",
			oldVersion:   "18.0.0",
			expectedCidr: "192.168.0.0/16",
		},
		{
			name: "case 4: upgrade from v17 to v18 with networkpool set",
			ctx:  context.Background(),

			newVersion:   "17.0.0",
			oldVersion:   "18.0.0",
			networkPool:  true,
			expectedCidr: "",
		},
		{
			name: "case 5: upgrade from v17 to v18 with custom pod cidr",
			ctx:  context.Background(),

			newVersion:    "17.0.0",
			oldVersion:    "18.0.0",
			customPodCidr: true,
			expectedCidr:  "",
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}
			// create releases
			releases := []releasev1alpha1.Release{
				unittest.NamedRelease("v17.0.0"),
				unittest.NamedRelease("v18.0.0"),
			}
			for _, release := range releases {
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, &release)
				if err != nil {
					t.Fatal(err)
				}
			}

			awsCluster := unittest.DefaultAWSCluster()
			if tc.networkPool {
				awsCluster.Spec.Provider.Nodes.NetworkPool = "np1"
			}
			if tc.customPodCidr {
				awsCluster.Spec.Provider.Pods.CIDRBlock = "10.199.0.0/16"
			}
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsCluster)
			if err != nil {
				t.Fatal(err)
			}

			// create old and new objects
			cluster := unittest.DefaultCluster()
			oldCluster := unittest.DefaultCluster()
			cluster.SetLabels(map[string]string{label.Release: tc.newVersion})
			oldCluster.SetLabels(map[string]string{label.Release: tc.oldVersion})

			// run mutate function to default cluster operator label
			var patch []mutator.PatchOperation
			oldV := semver.MustParse(tc.oldVersion)
			newV := semver.MustParse(tc.newVersion)

			patch, err = mutate.DefaultCiliumCidrOnV18Upgrade(*cluster, &oldV, &newV)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			var cidr string
			for _, p := range patch {
				if p.Path == "/metadata/annotations" {
					annotations := p.Value.(map[string]string)

					cidr = annotations[annotation.CiliumPodCidr]
				}
			}
			// check if the release label is as expected
			if tc.expectedCidr != cidr {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedCidr, cidr)
			}
		})
	}
}
