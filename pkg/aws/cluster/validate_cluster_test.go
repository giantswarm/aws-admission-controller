package cluster

import (
	"context"
	"strconv"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/unittest"
)

func TestValidateReleaseVersion(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		oldReleaseVersion string
		newReleaseVersion string
		valid             bool
	}{
		{
			// Version unchanged
			name: "case 0",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "3.0.0",
			valid:             true,
		},
		{
			// version changed to valid release
			name: "case 1",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "4.0.0",
			valid:             true,
		},
		{
			// version changed to deprecated release
			name: "case 2",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "3.2.0",
			valid:             false,
		},
		{
			// version changed to invalid release
			name: "case 3",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "3.3.0",
			valid:             false,
		},
		{
			// version changed with major skip
			name: "case 4",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "5.0.0",
			valid:             false,
		},
		{
			// version changed to older release
			name: "case 5",
			ctx:  context.Background(),

			oldReleaseVersion: "3.0.0",
			newReleaseVersion: "2.0.0",
			valid:             false,
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			handle := &Validator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			// create releases for testing
			releases := []unittest.ReleaseData{
				{
					Name:  "v4.0.0",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v3.2.0",
					State: releasev1alpha1.StateDeprecated,
				},
				{
					Name:  "v5.0.0",
					State: releasev1alpha1.StateActive,
				},
				{
					Name:  "v2.0.0",
					State: releasev1alpha1.StateActive,
				},
			}
			for _, r := range releases {
				release := unittest.DefaultRelease()
				release.SetName(r.Name)
				release.Spec.State = r.State
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, &release)
				if err != nil {
					t.Fatal(err)
				}
			}

			// create old and new object with release version labels
			oldObject := unittest.DefaultCluster()
			oldLabels := unittest.DefaultLabels()
			oldLabels[label.ReleaseVersion] = tc.oldReleaseVersion
			oldObject.SetLabels(oldLabels)

			newObject := unittest.DefaultCluster()
			newLabels := unittest.DefaultLabels()
			newLabels[label.ReleaseVersion] = tc.newReleaseVersion
			newObject.SetLabels(newLabels)

			// check if the result is as expected
			err = handle.ReleaseVersionValid(&oldObject, &newObject)
			if tc.valid && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error but returned %v", err)
			}
		})
	}
}
