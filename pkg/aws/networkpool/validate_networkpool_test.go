package networkpool

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
)

func TestNetworkPool(t *testing.T) {
	testCases := []struct {
		name                     string
		ctx                      context.Context
		customNetworkCIDR        string
		dockerCIDR               string
		kubernetesClusterIPRange string
		networkPoolCIDRs         []string
		tenantNetworkCIDR        string

		allowed bool
	}{
		{
			// No intersection
			name:                     "case 0",
			ctx:                      context.Background(),
			customNetworkCIDR:        "192.168.178.0/24",
			dockerCIDR:               "172.18.224.1/19",
			kubernetesClusterIPRange: "10.35.0.0/17",
			networkPoolCIDRs:         []string{"192.168.179.0/24", "172.16.0.0/16"},
			tenantNetworkCIDR:        "10.0.0.0/16",

			allowed: true,
		},
		{
			// Intersection
			name:                     "case 1",
			ctx:                      context.Background(),
			customNetworkCIDR:        "172.16.0.0/16",
			dockerCIDR:               "172.18.224.1/19",
			kubernetesClusterIPRange: "10.35.0.0/17",
			networkPoolCIDRs:         []string{"172.16.1.0/20"},
			tenantNetworkCIDR:        "10.0.0.0/16",

			allowed: false,
		},
		{
			// Intersection
			name:                     "case 2",
			ctx:                      context.Background(),
			customNetworkCIDR:        "10.0.16.0/16",
			dockerCIDR:               "172.18.224.1/19",
			kubernetesClusterIPRange: "10.35.0.0/17",
			networkPoolCIDRs:         []string{"172.16.0.0/16"},
			tenantNetworkCIDR:        "10.0.0.0/8",

			allowed: false,
		},
		{
			// Intersection
			name:                     "case 3",
			ctx:                      context.Background(),
			customNetworkCIDR:        "10.0.255.0/16",
			dockerCIDR:               "172.18.224.1/19",
			kubernetesClusterIPRange: "10.35.0.0/17",
			networkPoolCIDRs:         []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
			tenantNetworkCIDR:        "10.0.0.0/24",

			allowed: false,
		},
		{
			// No intersection
			name:                     "case 4",
			ctx:                      context.Background(),
			customNetworkCIDR:        "172.16.0.0/16",
			dockerCIDR:               "172.18.224.1/19",
			kubernetesClusterIPRange: "10.35.0.0/17",
			tenantNetworkCIDR:        "10.0.0.0/16",

			allowed: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				dockerCIDR:               tc.dockerCIDR,
				ipamNetworkCIDR:          tc.tenantNetworkCIDR,
				k8sClient:                fakeK8sClient,
				kubernetesClusterIPRange: tc.kubernetesClusterIPRange,
				logger:                   microloggertest.New(),
			}

			// create NetworkPools
			for _, networkPoolCIDR := range tc.networkPoolCIDRs {
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, unittest.DefaultNetworkPool(networkPoolCIDR))
				if err != nil {
					t.Fatal(err)
				}
			}

			// simulate a admission request for NetworkPool creation
			request, err := unittest.DefaultAdmissionRequestNetworkPool(tc.customNetworkCIDR)
			if err != nil {
				t.Fatal(err)
			}
			allowed, err := validate.Validate(&request)
			if tc.allowed != allowed {
				t.Fatalf("expected %v to not to differ from %v: %v", allowed, tc.allowed, err)
			}
		})
	}
}
