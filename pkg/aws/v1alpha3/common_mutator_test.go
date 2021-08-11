package v1alpha3

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-admission-controller/v3/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v3/pkg/unittest/v1alpha3"
)

func TestLabelFromCluster(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentRelease string
		expectedPatch  string
	}{
		{
			// Don't default the Release Label if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentRelease: unittest.DefaultReleaseVersion,
			expectedPatch:  "",
		},
		{
			// Default the Release Label if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentRelease: "",
			expectedPatch:  unittest.DefaultReleaseVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedRelease string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			// run mutate function to default AWSCluster ReleaseVersion label
			var patch []mutator.PatchOperation
			cluster := unittest.DefaultCluster()
			awscluster := unittest.DefaultAWSCluster()
			awscluster.SetLabels(map[string]string{label.Release: tc.currentRelease, label.Cluster: unittest.DefaultClusterID})
			patch, err = MutateLabelFromCluster(mutate, awscluster.GetObjectMeta(), *cluster, label.Release)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.Release)) {
					updatedRelease = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedRelease {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedRelease)
			}
		})
	}
}

func TestLabelFromAWSCluster(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentOperator string
		expectedPatch   string
	}{
		{
			// Don't default the Operator Label if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentOperator: unittest.DefaultAWSOperatorVersion,
			expectedPatch:   "",
		},
		{
			// Default the Operator Label if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentOperator: "",
			expectedPatch:   unittest.DefaultAWSOperatorVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedOperator string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			// run mutate function to default AWSControlplane operator label
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			awscontrolplane := unittest.DefaultAWSControlPlane()
			awscontrolplane.SetLabels(map[string]string{label.AWSOperatorVersion: tc.currentOperator, label.Cluster: unittest.DefaultClusterID})
			patch, err = MutateLabelFromAWSCluster(mutate, awscontrolplane.GetObjectMeta(), awscluster, label.AWSOperatorVersion)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.AWSOperatorVersion)) {
					updatedOperator = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedOperator {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedOperator)
			}
		})
	}
}

func TestLabelFromRelease(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentOperator string
		expectedPatch   string
	}{
		{
			// Don't default the Operator Label if it is set
			name: "case 0",
			ctx:  context.Background(),

			currentOperator: unittest.DefaultAWSOperatorVersion,
			expectedPatch:   "",
		},
		{
			// Default the Operator Label if it is not set
			name: "case 1",
			ctx:  context.Background(),

			currentOperator: "",
			expectedPatch:   unittest.DefaultAWSOperatorVersion,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedOperator string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			// run mutate function to default AWSControlplane operator label
			var patch []mutator.PatchOperation
			awscluster := unittest.DefaultAWSCluster()
			release := unittest.DefaultRelease()
			awscluster.SetLabels(map[string]string{label.AWSOperatorVersion: tc.currentOperator, label.Release: unittest.DefaultReleaseVersion})
			patch, err = MutateLabelFromRelease(mutate, awscluster.GetObjectMeta(), release, label.AWSOperatorVersion, "aws-operator")
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == fmt.Sprintf("/metadata/labels/%s", EscapeJSONPatchString(label.AWSOperatorVersion)) {
					updatedOperator = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedOperator {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedOperator)
			}
		})
	}
}

func TestMutateNamespace(t *testing.T) {
	testCases := []struct {
		name             string
		currentNamespace string
		organization     string
		releaseVersion   string

		expectedPatch string
	}{
		{
			// Don't default the Namespace if it is set
			name:             "case 0",
			currentNamespace: "org-test",
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "18.0.0",

			expectedPatch: "",
		},
		{
			// Default the Namespace if it is not set
			name:             "case 1",
			currentNamespace: metav1.NamespaceDefault,
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "18.0.0",

			expectedPatch: key.OrganizationNamespaceFromName(unittest.DefaultOrganizationName),
		},
		{
			// Default the Namespace if it is not set
			name:             "case 2",
			currentNamespace: "",
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "18.0.0",

			expectedPatch: key.OrganizationNamespaceFromName(unittest.DefaultOrganizationName),
		},
		{
			// Don't default the Namespace in older version
			name:             "case 3",
			currentNamespace: metav1.NamespaceDefault,
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "14.0.0",

			expectedPatch: "",
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedNamespace string
			var patch []mutator.PatchOperation

			awscluster := unittest.DefaultAWSCluster()
			awscluster.SetLabels(map[string]string{label.Organization: tc.organization, label.Release: tc.releaseVersion})
			awscluster.SetNamespace(tc.currentNamespace)
			releaseVersion, err := ReleaseVersion(&awscluster, patch)
			if err != nil {
				t.Fatal(err)
			}

			// run mutate function to default AWSCluster Namespace
			mutate := &Handler{
				K8sClient: unittest.FakeK8sClient(),
				Logger:    microloggertest.New(),
			}
			patch, err = MutateNamespace(mutate, awscluster.GetObjectMeta(), releaseVersion)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/metadata/namespace" {
					updatedNamespace = p.Value.(string)
				}
			}
			// check if the release label is as expected
			if tc.expectedPatch != updatedNamespace {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedPatch, updatedNamespace)
			}
		})
	}
}
