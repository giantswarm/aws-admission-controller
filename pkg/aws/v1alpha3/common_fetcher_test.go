package v1alpha3

import (
	"context"
	"strconv"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
)

func TestNewestReleaseVersion(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		releases        []unittest.ReleaseData
		expectedVersion string
	}{
		{
			// take only active releases
			name: "case 0",
			ctx:  context.Background(),

			releases: []unittest.ReleaseData{
				{
					Name:  "v1.2.3",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v2.1.0",
					State: releasev1alpha1.StateDeprecated,
				},
			},
			expectedVersion: "1.2.3",
		},
		{
			// don't take dev versions
			name: "case 1",
			ctx:  context.Background(),

			releases: []unittest.ReleaseData{
				{
					Name:  "v1.2.3",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v3.2.3-dev",
					State: releasev1alpha1.StateActive,
				},
			},
			expectedVersion: "1.2.3",
		},
		{
			// sort releases correctly
			name: "case 1",
			ctx:  context.Background(),

			releases: []unittest.ReleaseData{
				{
					Name:  "v1.1.1",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v0.2.3",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v1.2.3",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v1.2.1",
					State: releasev1alpha1.StateActive,
				},
			},
			expectedVersion: "1.2.3",
		},
		{
			// don't take CAPI versions
			name: "case 1",
			ctx:  context.Background(),

			releases: []unittest.ReleaseData{
				{
					Name:  "v1.2.3",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v20.0.0-v1alpha3",
					State: releasev1alpha1.StateActive,
				},
			},
			expectedVersion: "1.2.3",
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
			// create releases for testing
			for _, r := range tc.releases {
				release := unittest.DefaultRelease()
				release.SetName(r.Name)
				release.Spec.State = r.State
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, &release)
				if err != nil {
					t.Fatal(err)
				}
			}
			// run fetcher to get newest active release version
			version, err := FetchNewestReleaseVersion(handle)
			if err != nil {
				t.Fatal(err)
			}
			// check if the release version is as expected
			if tc.expectedVersion != version.String() {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedVersion, version.String())
			}
		})
	}
}
