package azureupdate

import (
	"context"
	"encoding/json"
	"testing"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/admission-controller/pkg/unittest"
)

var (
	controlPlaneName      = "gmk24"
	controlPlaneNameSpace = "default"
)

func TestAzureConfigValidate(t *testing.T) {
	releases := []string{"11.3.0", "11.3.1", "11.4.0", "12.0.0"}

	testCases := []struct {
		name         string
		ctx          context.Context
		releases     []string
		oldVersion   string
		newVersion   string
		conditions   []string
		allowed      bool
		errorMatcher func(err error) bool
	}{
		{
			name: "case 0",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.3.1",
			conditions:   []string{},
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name: "case 1",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.4.0",
			conditions:   []string{},
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name: "case 2",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "12.0.0",
			conditions:   []string{},
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name: "case 3",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.0",
			newVersion:   "11.3.0",
			conditions:   []string{},
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name: "case 4",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "11.4.0",
			conditions:   []string{},
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name: "case 5",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "",
			conditions:   []string{},
			allowed:      false,
			errorMatcher: IsParsingFailed,
		},
		{
			name: "case 6",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "",
			newVersion:   "11.3.1",
			conditions:   []string{},
			allowed:      false,
			errorMatcher: IsParsingFailed,
		},
		{
			name: "case 7",
			ctx:  context.Background(),

			releases:     []string{"invalid"},
			oldVersion:   "11.3.0",
			newVersion:   "11.4.0",
			conditions:   []string{},
			allowed:      false,
			errorMatcher: IsInvalidReleaseError,
		},
		{
			name: "case 8",
			ctx:  context.Background(),

			releases:     []string{"invalid"},
			oldVersion:   "11.3.0",
			newVersion:   "11.3.1",
			conditions:   []string{},
			allowed:      false,
			errorMatcher: IsInvalidReleaseError,
		},
		{
			name: "case 9",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.3.1",
			newVersion:   "11.3.0",
			conditions:   []string{},
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name: "case 10",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.0.0", // does not exist
			newVersion:   "11.3.0", // exists
			conditions:   []string{},
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name: "case 11",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0", // exists
			newVersion:   "11.5.0", // does not exist
			conditions:   []string{},
			allowed:      false,
			errorMatcher: IsInvalidReleaseError,
		},
		{
			name: "case 12",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.5.0", // does not exist
			newVersion:   "11.5.0", // does not exist
			conditions:   []string{},
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name: "case 13",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0",
			newVersion:   "11.4.0",
			conditions:   []string{},
			allowed:      true,
			errorMatcher: nil,
		},
		{
			name: "case 14",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0",
			newVersion:   "11.4.1",
			conditions:   []string{conditionCreating},
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
		{
			name: "case 15",
			ctx:  context.Background(),

			releases:     releases,
			oldVersion:   "11.4.0",
			newVersion:   "11.4.1",
			conditions:   []string{conditionUpdating},
			allowed:      false,
			errorMatcher: IsInvalidOperationError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Create a new logger that is used by all admitters.
			var newLogger micrologger.Logger
			{
				newLogger, err = micrologger.New(micrologger.Config{})
				if err != nil {
					panic(microerror.JSON(err))
				}
			}
			fakeK8sClient := unittest.FakeK8sClient()
			admit := &AzureConfigAdmitter{
				k8sClient: fakeK8sClient,
				logger:    newLogger,
			}

			// Create needed releases.
			err = ensureReleases(fakeK8sClient.G8sClient(), tc.releases)
			if err != nil {
				t.Fatal(err)
			}

			// Run admission request to validate AzureConfig updates.
			allowed, err := admit.Validate(getAdmissionRequest(tc.oldVersion, tc.newVersion, tc.conditions))

			// Check if the error is the expected one.
			switch {
			case err == nil && tc.errorMatcher == nil:
				// fall through
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("expected %#v got %#v", nil, err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("expected %#v got %#v", "error", nil)
			case !tc.errorMatcher(err):
				t.Fatalf("unexpected error: %#v", err)
			}

			// Check if the validation result is the expected one.
			if tc.allowed != allowed {
				t.Fatalf("expected %v to be equal to %v", tc.allowed, allowed)
			}
		})
	}
}

func getAdmissionRequest(oldVersion string, newVersion string, clusterConditionTypes []string) *v1beta1.AdmissionRequest {
	req := &v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AzureClusterUpgrade",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "provider.giantswarm.io/v1alpha1",
			Resource: "azureconfigs",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    azureConfigRawObj(newVersion, []string{}),
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    azureConfigRawObj(oldVersion, clusterConditionTypes),
			Object: nil,
		},
	}

	return req
}

func azureConfigRawObj(version string, clusterConditionTypes []string) []byte {
	var conditions []providerv1alpha1.StatusClusterCondition
	for _, cond := range clusterConditionTypes {
		conditions = append(conditions, providerv1alpha1.StatusClusterCondition{Type: cond})
	}
	azureconfig := providerv1alpha1.AzureConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureConfig",
			APIVersion: "provider.giantswarm.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/control-plane":   controlPlaneName,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": version,
			},
		},
		Spec: providerv1alpha1.AzureConfigSpec{
			Cluster:       providerv1alpha1.Cluster{},
			Azure:         providerv1alpha1.AzureConfigSpecAzure{},
			VersionBundle: providerv1alpha1.AzureConfigSpecVersionBundle{},
		},
		Status: providerv1alpha1.AzureConfigStatus{
			Cluster: providerv1alpha1.StatusCluster{
				Conditions: conditions,
			},
		},
	}
	byt, _ := json.Marshal(azureconfig)
	return byt
}
