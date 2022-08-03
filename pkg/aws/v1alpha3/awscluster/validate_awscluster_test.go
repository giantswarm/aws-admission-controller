package awscluster

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"

	unittest "github.com/giantswarm/aws-admission-controller/v4/pkg/unittest/v1alpha3"
)

func TestAWSCNIPrefix(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		region string
		err    error
	}{
		{
			// CNI prefix is allowed.
			name: "case 0",
			ctx:  context.Background(),

			region: "us-west-2",
			err:    nil,
		},
		{
			// CNI prefix is not allowed.
			name: "case 1",
			ctx:  context.Background(),

			region: "cn-north-1",
			err:    notAllowedError,
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

			// run admission request to default AWSCluster Pod CIDR
			awsCluster := unittest.DefaultAWSCluster()

			// set CNI prefix annotation
			awsCluster.SetAnnotations(map[string]string{
				annotation.AWSCNIPrefixDelegation: "true",
			})

			// set region
			awsCluster.Spec.Provider.Region = tc.region

			err = validate.AWSClusterAnnotationCNIPrefix(*awsCluster)
			if microerror.Cause(err) != tc.err {
				t.Fatal(err)
			}
		})
	}
}

func TestCilium(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		ciliumCidr    string
		ipamCidrBlock string
		podCidrBlock  string
		err           error
	}{
		{
			// CNI CIDR is not allowed too small.
			name: "case 0",
			ctx:  context.Background(),

			ciliumCidr:    "10.0.0.0/25",
			ipamCidrBlock: "10.5.0.0/16",
			podCidrBlock:  "10.4.0.0/16",
			err:           notAllowedError,
		},
		{
			// CNI CIDR is not allowed, overlapping CIDR's.
			name: "case 1",
			ctx:  context.Background(),

			ciliumCidr:    "10.0.0.0/8",
			ipamCidrBlock: "10.5.0.0/16",
			podCidrBlock:  "10.0.0.0/16",
			err:           notAllowedError,
		},
		{
			// CNI CIDR is not allowed, overlapping CIDR's.
			name: "case 1",
			ctx:  context.Background(),

			ciliumCidr:    "10.0.0.0/16",
			ipamCidrBlock: "10.1.0.0/16",
			podCidrBlock:  "10.2.0.0/16",
			err:           nil,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			validate := &Validator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),

				ipamCidrBlock: tc.ipamCidrBlock,
				podCIDRBlock:  tc.podCidrBlock,
			}

			// run admission request to default AWSCluster Pod CIDR
			awsCluster := unittest.DefaultAWSCluster()

			// set CNI prefix annotation
			awsCluster.SetAnnotations(map[string]string{
				annotation.CiliumPodCidr: tc.ciliumCidr,
			})

			err = validate.Cilium(*awsCluster)
			if microerror.Cause(err) != tc.err {
				t.Fatal(err)
			}
		})
	}
}
