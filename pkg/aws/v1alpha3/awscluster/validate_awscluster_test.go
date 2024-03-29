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

func Test_AWSClusterAlreadyExists(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
	}{
		{
			name:  "case 0: cluster does not exists",
			valid: true,
		},
		{
			name:  "case 1: cluster already exists",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := &Validator{
				k8sClient: unittest.FakeK8sClient(),
				logger:    microloggertest.New(),
			}
			awsCluster := unittest.DefaultAWSCluster()

			if !tc.valid {
				// create the same AWS cluster in giantswarm namespace
				awsCluster.Namespace = "giantswarm"
				err := v.k8sClient.CtrlClient().Create(context.TODO(), awsCluster)
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
			}

			// check if the result is as expected
			err := v.AWSClusterExists(awsCluster)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}
