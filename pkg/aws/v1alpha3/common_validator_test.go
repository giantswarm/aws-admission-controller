package v1alpha3

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v3/pkg/key"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
	unittest "github.com/giantswarm/aws-admission-controller/v3/pkg/unittest/v1alpha3"

	securityv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/security/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_MaxBatchSizeIsValid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "case 0: int - simple value",
			input: "5",
			valid: true,
		},
		{
			name:  "case 1: int - big value",
			input: "200",
			valid: true,
		},
		{
			name:  "case 2: int - invalid value - negative number",
			input: "-10",
			valid: false,
		},
		{
			name:  "case 2: int - invalid value - zero",
			input: "0",
			valid: false,
		},
		{
			name:  "case 4: percentage - simple value",
			input: "0.5",
			valid: true,
		},
		{
			name:  "case 5: percentage - rounding",
			input: "0.35",
			valid: true,
		},
		{
			name:  "case 6: percentage - rounding",
			input: "0.32",
			valid: true,
		},
		{
			name:  "case 7: percentage - invalid value - too big",
			input: "1.5",
			valid: false,
		},
		{
			name:  "case 8: percentage - invalid value - negative",
			input: "-0.5",
			valid: false,
		},
		{
			name:  "case 9: invalid value - '50%'",
			input: "50%",
			valid: false,
		},
		{
			name:  "case 10: invalid value - string",
			input: "test",
			valid: false,
		},
		{
			name:  "case 11: invalid value - number and string",
			input: "5erft",
			valid: false,
		},
		{
			name:  "case 12: invalid value - float and string",
			input: "0.5erft",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			isValid := MaxBatchSizeIsValid(tc.input)

			if isValid != tc.valid {
				t.Fatalf("%s - expected '%t' got '%t'\n", tc.name, tc.valid, isValid)
			}
		})
	}
}

func Test_PauseTimeIsValid(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "case 0: simple value",
			value: "PT15M",
			valid: true,
		},
		{
			name:  "case 2: simple value",
			value: "PT10S",
			valid: true,
		},
		{
			name:  "case 3: simple value",
			value: "PT2M10S",
			valid: true,
		},
		{
			name:  "case 4: simple value",
			value: "PT2M10S",
			valid: true,
		},
		{
			name:  "case 5: invalid value value",
			value: "10m",
			valid: false,
		},
		{
			name:  "case 6: invalid value value",
			value: "10s",
			valid: false,
		},
		{
			name:  "case 7: invalid value value",
			value: "10",
			valid: false,
		},
		{
			name:  "case 8: invalid value value",
			value: "1 hour",
			valid: false,
		},
		{
			name:  "case 9: invalid value value",
			value: "random string",
			valid: false,
		},
		{
			name:  "case 10: invalid value value",
			value: "",
			valid: false,
		},
		{
			name:  "case 11: duration too big",
			value: "PT1H2M",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := PauseTimeIsValid(tc.value)

			if result != tc.valid {
				t.Fatalf("%s -  expected '%t' got '%t'\n", tc.name, tc.valid, result)
			}
		})
	}
}

func TestValidateLabelKeys(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		newLabels map[string]string
		valid     bool
	}{
		{
			// no label key was changed
			name: "case 0",
			ctx:  context.Background(),

			newLabels: unittest.DefaultLabels(),
			valid:     true,
		},
		{
			// non giantswarm label key was changed
			name: "case 1",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                  unittest.DefaultClusterID,
				label.ClusterOperatorVersion:   unittest.DefaultClusterOperatorVersion,
				label.Release:                  "100.0.0",
				label.Organization:             "example-organization",
				"example-key-changed":          "example-value",
				unittest.DefaultProviderTagKey: unittest.DefaultProviderTagValue,
			},
			valid: true,
		},
		{
			// giantswarm label key was changed
			name: "case 2",
			ctx:  context.Background(),

			newLabels: map[string]string{
				"giantswarm.io/cluster-changed": unittest.DefaultClusterID,
				label.ClusterOperatorVersion:    unittest.DefaultClusterOperatorVersion,
				label.Release:                   unittest.DefaultReleaseVersion,
				label.Organization:              "example-organization",
				unittest.DefaultProviderTagKey:  unittest.DefaultProviderTagValue,
			},
			valid: false,
		},
		{
			// provider tag label key was changed
			name: "case 3",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                unittest.DefaultClusterID,
				label.ClusterOperatorVersion: unittest.DefaultClusterOperatorVersion,
				label.Release:                unittest.DefaultReleaseVersion,
				label.Organization:           "example-organization",
				"tag.provider.giantswarm.io/NewTaggingVersion": unittest.DefaultProviderTagValue,
			},
			valid: true,
		},
		{
			// provider tag label was removed
			name: "case 4",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                unittest.DefaultClusterID,
				label.ClusterOperatorVersion: unittest.DefaultClusterOperatorVersion,
				label.Release:                unittest.DefaultReleaseVersion,
				label.Organization:           "example-organization",
			},
			valid: true,
		},
		{
			// provider tag label was added
			name: "case 5",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                                  unittest.DefaultClusterID,
				label.ClusterOperatorVersion:                   unittest.DefaultClusterOperatorVersion,
				label.Release:                                  unittest.DefaultReleaseVersion,
				label.Organization:                             "example-organization",
				unittest.DefaultProviderTagKey:                 unittest.DefaultProviderTagValue,
				"tag.provider.giantswarm.io/NewTaggingVersion": "14",
			},
			valid: true,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			handle := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			oldObject := unittest.DefaultCluster()
			newObject := unittest.DefaultCluster()
			newObject.SetLabels(tc.newLabels)
			err = ValidateLabelKeys(handle, oldObject, newObject)
			// check if the result is as expected
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func TestValidateLabelValues(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		newLabels map[string]string
		valid     bool
	}{
		{
			// no label was changed
			name: "case 0",
			ctx:  context.Background(),

			newLabels: unittest.DefaultLabels(),
			valid:     true,
		},
		{
			// version label was changed
			name: "case 1",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                  unittest.DefaultClusterID,
				label.ClusterOperatorVersion:   "1.2.3",
				label.Release:                  "100.0.0",
				label.Organization:             "example-organization",
				unittest.DefaultProviderTagKey: unittest.DefaultProviderTagValue,
			},
			valid: true,
		},
		{
			// label was changed
			name: "case 2",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                  unittest.DefaultClusterID,
				label.ClusterOperatorVersion:   unittest.DefaultClusterOperatorVersion,
				label.Release:                  unittest.DefaultReleaseVersion,
				label.Organization:             "new-organization",
				unittest.DefaultProviderTagKey: unittest.DefaultProviderTagValue,
			},
			valid: false,
		},
		{
			// provider tag label was changed
			name: "case 3",
			ctx:  context.Background(),

			newLabels: map[string]string{
				label.Cluster:                  unittest.DefaultClusterID,
				label.ClusterOperatorVersion:   "1.2.3",
				label.Release:                  "0.0.0",
				label.Organization:             "example-organization",
				unittest.DefaultProviderTagKey: "v2.4",
			},
			valid: true,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			handle := &Handler{
				K8sClient: fakeK8sClient,
				Logger:    microloggertest.New(),
			}
			oldObject := unittest.DefaultCluster()
			newObject := unittest.DefaultCluster()
			newObject.SetLabels(tc.newLabels)
			err = ValidateLabelValues(handle, oldObject, newObject)
			// check if the result is as expected
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}

func Test_Organization(t *testing.T) {
	testCases := []struct {
		ctx   context.Context
		name  string
		input string
	}{
		{
			ctx:   context.Background(),
			name:  "case 0",
			input: "new-organization",
		},
		{
			ctx:   context.Background(),
			name:  "case 1",
			input: "different-organization",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			// create fake organizations
			organizations := []*securityv1alpha1.Organization{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: tc.input,
					},
					Spec: securityv1alpha1.OrganizationSpec{},
				},
				// this is the organization label which is from the default test cluster
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-organization",
					},
					Spec: securityv1alpha1.OrganizationSpec{},
				},
			}

			fakeK8sClient := unittest.FakeK8sClient()
			for _, org := range organizations {
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, org)
				if err != nil {
					panic(err)
				}
			}

			// fail if organization is not found
			err = ValidateOrganizationLabelContainsExistingOrganization(tc.ctx, fakeK8sClient.CtrlClient(), unittest.DefaultCluster())
			if IsOrganizationNotFoundError(err) {
				t.Fatalf("%s: it should fail when using a non existing organization", tc.name)
			}

			// fail if organization label is empty
			err = ValidateOrganizationLabelContainsExistingOrganization(tc.ctx, fakeK8sClient.CtrlClient(), unittest.DefaultClusterEmptyOrganization())
			if !IsOrganizationNotFoundError(err) {
				t.Fatalf("%s: it should always fail if organization label is empty", tc.name)
			}

			// fail if organization label is not present
			err = ValidateOrganizationLabelContainsExistingOrganization(tc.ctx, fakeK8sClient.CtrlClient(), unittest.DefaultClusterWithoutOrganizationLabel())
			if !IsOrganizationLabelNotFoundError(err) {
				t.Fatalf("%s: it should always fail if organization label does not exist", tc.name)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	testCases := []struct {
		name             string
		currentNamespace string
		organization     string
		releaseVersion   string

		valid bool
	}{
		{
			// invalid namespace
			name:             "case 0",
			currentNamespace: "org-test",
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "18.0.0",

			valid: false,
		},
		{
			// Default Namespace
			name:             "case 1",
			currentNamespace: metav1.NamespaceDefault,
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "18.0.0",

			valid: false,
		},
		{
			// no Namespace
			name:             "case 2",
			currentNamespace: "",
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "18.0.0",

			valid: false,
		},
		{
			// Don't validate the Namespace in older version
			name:             "case 3",
			currentNamespace: metav1.NamespaceDefault,
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "14.0.0",

			valid: true,
		},
		{
			// Valid namespace
			name:             "case 4",
			currentNamespace: key.OrganizationNamespaceFromName(unittest.DefaultOrganizationName),
			organization:     unittest.DefaultOrganizationName,
			releaseVersion:   "18.0.0",

			valid: true,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			awscluster := unittest.DefaultAWSCluster()
			awscluster.SetLabels(map[string]string{label.Organization: tc.organization, label.Release: tc.releaseVersion})
			awscluster.SetNamespace(tc.currentNamespace)

			// validate AWSCluster Namespace
			err = ValidateNamespace(awscluster.GetObjectMeta())

			// check if the result is as expected
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}
