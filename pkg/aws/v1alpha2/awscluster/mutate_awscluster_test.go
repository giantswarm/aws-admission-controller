package awscluster

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/apimachinery/pkg/types"

	awsv1alpha2 "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v2/pkg/unittest/v1alpha2"
)

func TestAWSClusterPodCIDR(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentPodCIDR  string
		expectedPodCIDR string
	}{
		{
			// Don't default the Pod CIDR if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentPodCIDR:  unittest.DefaultPodCIDR,
			expectedPodCIDR: "",
		},
		{
			// Default the Pod CIDR if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentPodCIDR:  "",
			expectedPodCIDR: unittest.DefaultPodCIDR,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedCIDR map[string]string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				podCIDRBlock: unittest.DefaultPodCIDR,
				k8sClient:    fakeK8sClient,
				logger:       microloggertest.New(),
			}

			// run admission request to default AWSCluster Pod CIDR
			var patch []mutator.PatchOperation
			request, err := unittest.CustomAdmissionRequestAWSCluster(tc.currentPodCIDR)
			if err != nil {
				t.Fatal(err)
			}
			patch, err = mutate.Mutate(&request)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/provider/pods" {
					updatedCIDR = p.Value.(map[string]string)
				}
			}
			// check if the pod CIDR is as expected
			if tc.expectedPodCIDR != updatedCIDR["cidrBlock"] {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPodCIDR, updatedCIDR)
			}
		})
	}
}
func TestAWSClusterCredentials(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentCredential types.NamespacedName
		secretExists      bool
		expectedPatch     types.NamespacedName
	}{
		{
			// Don't default the Credential if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentCredential: unittest.DefaultClusterCredentialSecretLocation(),
			expectedPatch:     types.NamespacedName{},
		},
		{
			// Default the Credential if it is not set and no org credential secret exists
			name: "case 1",
			ctx:  context.Background(),

			currentCredential: types.NamespacedName{},
			expectedPatch:     awsv1alpha2.DefaultCredentialSecret(),
		},
		{
			// Default the Credential if it is not set and an org credential secret exists
			name: "case 2",
			ctx:  context.Background(),

			currentCredential: types.NamespacedName{},
			secretExists:      true,
			expectedPatch:     unittest.DefaultClusterCredentialSecretLocation(),
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedCredential map[string]string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}
			if tc.secretExists {
				secret := unittest.DefaultClusterCredentialSecret()

				err = fakeK8sClient.CtrlClient().Create(tc.ctx, &secret)
				if err != nil {
					t.Fatal(err)
				}
			}
			// run mutate function to default AWSCluster Credential
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			awscluster.Spec.Provider.CredentialSecret.Name = tc.currentCredential.Name
			awscluster.Spec.Provider.CredentialSecret.Namespace = tc.currentCredential.Namespace
			patch, err = mutate.MutateCredential(awscluster)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/provider/credentialSecret" {
					updatedCredential = p.Value.(map[string]string)
				}
			}
			// check if the pod CIDR is as expected
			if tc.expectedPatch.Name != updatedCredential["name"] || tc.expectedPatch.Namespace != updatedCredential["namespace"] {
				t.Fatalf("expected %#q/%#q to be equal to %#q/%#q",
					tc.expectedPatch.Namespace,
					tc.expectedPatch.Name,
					updatedCredential["namespace"],
					updatedCredential["name"])
			}
		})
	}
}

func TestAWSClusterDescription(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentDescription string
		expectedPatch      string
	}{
		{
			// Don't default the Cluster Description if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentDescription: "My cluster",
			expectedPatch:      "",
		},
		{
			// Default the Cluster Description if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentDescription: "",
			expectedPatch:      awsv1alpha2.DefaultClusterDescription,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedDescription string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				podCIDRBlock: unittest.DefaultPodCIDR,
				k8sClient:    fakeK8sClient,
				logger:       microloggertest.New(),
			}

			// run mutate function to default AWSCluster Description
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			awscluster.Spec.Cluster.Description = tc.currentDescription
			patch, err = mutate.MutateDescription(awscluster)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/cluster/description" {
					updatedDescription = p.Value.(string)
				}
			}
			// check if the pod CIDR is as expected
			if tc.expectedPatch != updatedDescription {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedDescription)
			}
		})
	}
}
func TestAWSClusterDomain(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentDomain string
		expectedPatch string
	}{
		{
			// Don't default the Cluster DNS Domain if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentDomain: unittest.DefaultClusterDNSDomain,
			expectedPatch: "",
		},
		{
			// Default the Cluster DNS Domain if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentDomain: "",
			expectedPatch: unittest.DefaultClusterDNSDomain,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedDomain string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				dnsDomain: unittest.DefaultClusterDNSDomain,
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			// run mutate function to default AWSCluster Description
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			awscluster.Spec.Cluster.DNS.Domain = tc.currentDomain
			patch, err = mutate.MutateDomain(awscluster)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/cluster/dns/domain" {
					updatedDomain = p.Value.(string)
				}
			}
			// check if the pod CIDR is as expected
			if tc.expectedPatch != updatedDomain {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedDomain)
			}
		})
	}
}

func TestAWSClusterMaster(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentAZ     string
		currentIT     string
		expectedPatch map[string]string
	}{
		{
			// Don't default the Master if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentAZ:     unittest.DefaultMasterAvailabilityZone,
			currentIT:     unittest.DefaultMasterInstanceType,
			expectedPatch: map[string]string{},
		},
		{
			// Default the Master if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentAZ: "",
			currentIT: "",
			expectedPatch: map[string]string{
				"availabilityZone": unittest.DefaultMasterAvailabilityZone,
				"instanceType":     awsv1alpha2.DefaultMasterInstanceType},
		},
		{
			// Default the Availability Zone if it is not set
			name: "case 2",
			ctx:  context.Background(),

			currentAZ: "",
			currentIT: unittest.DefaultMasterInstanceType,
			expectedPatch: map[string]string{
				"availabilityZone": unittest.DefaultMasterAvailabilityZone,
				"instanceType":     unittest.DefaultMasterInstanceType},
		},
		{
			// Default the Instance Type if it is not set
			name: "case 3",
			ctx:  context.Background(),

			currentAZ: "eu-central-1a",
			currentIT: "",
			expectedPatch: map[string]string{
				"availabilityZone": "eu-central-1a",
				"instanceType":     awsv1alpha2.DefaultMasterInstanceType},
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedMaster map[string]string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient:              fakeK8sClient,
				logger:                 microloggertest.New(),
				validAvailabilityZones: []string{unittest.DefaultMasterAvailabilityZone},
			}

			// run mutate function to default AWSCluster Master attributes
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			awscluster.Spec.Provider.Master.AvailabilityZone = tc.currentAZ
			awscluster.Spec.Provider.Master.InstanceType = tc.currentIT
			patch, err = mutate.MutateMasterPreHA(awscluster)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/provider/master" {
					updatedMaster = p.Value.(map[string]string)
				}
			}
			// check if the Master attribute is as expected
			if tc.expectedPatch["availabilityZone"] != updatedMaster["availabilityZone"] || tc.expectedPatch["instanceType"] != updatedMaster["instanceType"] {
				t.Fatalf("expected %#q/%#q to be equal to %#q/%#q",
					tc.expectedPatch["availabilityZone"],
					tc.expectedPatch["instanceType"],
					updatedMaster["availabilityZone"],
					updatedMaster["instanceType"])
			}
		})
	}
}

func TestAWSClusterRegion(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentRegion string
		expectedPatch string
	}{
		{
			// Don't default the Cluster Region if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentRegion: unittest.DefaultClusterRegion,
			expectedPatch: "",
		},
		{
			// Default the Cluster Region if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentRegion: "",
			expectedPatch: unittest.DefaultClusterRegion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedRegion string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				region:    unittest.DefaultClusterRegion,
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			// run mutate function to default AWSCluster Description
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			awscluster.Spec.Provider.Region = tc.currentRegion
			patch, err = mutate.MutateRegion(awscluster)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/provider/region" {
					updatedRegion = p.Value.(string)
				}
			}
			// check if the pod CIDR is as expected
			if tc.expectedPatch != updatedRegion {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedRegion)
			}
		})
	}
}
