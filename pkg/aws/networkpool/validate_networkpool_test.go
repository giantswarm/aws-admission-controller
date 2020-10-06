package networkpool

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestNetworkPool(t *testing.T) {
	testCases := []struct {
		name              string
		ctx               context.Context
		customNetworkCIDR string
		networkPoolCIDRs  []string
		tenantNetworkCIDR string

		allowed bool
	}{
		{
			// No intersection
			name:              "case 0",
			ctx:               context.Background(),
			customNetworkCIDR: "192.168.178.0/24",
			networkPoolCIDRs:  []string{"192.168.179.0/24", "172.16.0.0/16"},
			tenantNetworkCIDR: "10.0.0.0/16",

			allowed: true,
		},
		{
			// Intersection
			name:              "case 1",
			ctx:               context.Background(),
			tenantNetworkCIDR: "10.0.0.0/16",
			networkPoolCIDRs:  []string{"172.16.1.0/20"},
			customNetworkCIDR: "172.16.0.0/16",

			allowed: false,
		},
		{
			// Intersection
			name:              "case 2",
			ctx:               context.Background(),
			tenantNetworkCIDR: "10.0.0.0/8",
			networkPoolCIDRs:  []string{"172.16.0.0/16"},
			customNetworkCIDR: "10.0.16.0/16",

			allowed: false,
		},
		{
			// Intersection
			name:              "case 3",
			ctx:               context.Background(),
			tenantNetworkCIDR: "10.0.0.0/24",
			networkPoolCIDRs:  []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
			customNetworkCIDR: "10.0.255.0/16",

			allowed: false,
		},
		{
			// No intersection
			name:              "case 4",
			ctx:               context.Background(),
			customNetworkCIDR: "172.16.0.0/16",
			tenantNetworkCIDR: "10.0.0.0/16",

			allowed: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				ipamNetworkCIDR: tc.tenantNetworkCIDR,
				k8sClient:       fakeK8sClient,
				logger:          microloggertest.New(),
			}

			// create NetworkPools
			for _, networkPoolCIDR := range tc.networkPoolCIDRs {
				fakeK8sClient.CtrlClient().Create(tc.ctx, unittest.DefaultNetworkPool(networkPoolCIDR))
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
