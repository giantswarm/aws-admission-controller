package key

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1beta1"
)

func clusterWithLabels(labels map[string]string) *v1beta1.Cluster {
	return &v1beta1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
	}
}

func TestIsGitopsManaged(t *testing.T) {
	tests := []struct {
		name          string
		cluster       LabelsGetter
		wantNil       bool
		fluxNamespace string
		fluxName      string
	}{
		{
			name:          "No labels at all",
			cluster:       clusterWithLabels(nil),
			wantNil:       true,
			fluxNamespace: "",
			fluxName:      "",
		},
		{
			name:          "Only name label",
			cluster:       clusterWithLabels(map[string]string{"kustomize.toolkit.fluxcd.io/name": "test"}),
			wantNil:       true,
			fluxNamespace: "",
			fluxName:      "",
		},
		{
			name:          "Only namespace label",
			cluster:       clusterWithLabels(map[string]string{"kustomize.toolkit.fluxcd.io/namespace": "test"}),
			wantNil:       true,
			fluxNamespace: "",
			fluxName:      "",
		},
		{
			name: "Both labels",
			cluster: clusterWithLabels(map[string]string{
				"kustomize.toolkit.fluxcd.io/name":      "test",
				"kustomize.toolkit.fluxcd.io/namespace": "test",
			}),
			wantNil:       false,
			fluxNamespace: "test",
			fluxName:      "test",
		},
		{
			name:          "Only one unrelated label",
			cluster:       clusterWithLabels(map[string]string{"team": "phoenix"}),
			wantNil:       true,
			fluxNamespace: "",
			fluxName:      "",
		},
		{
			name: "Both labels but name with no value",
			cluster: clusterWithLabels(map[string]string{
				"kustomize.toolkit.fluxcd.io/name":      "",
				"kustomize.toolkit.fluxcd.io/namespace": "test",
			}),
			wantNil:       true,
			fluxNamespace: "",
			fluxName:      "",
		},
		{
			name: "Both labels but namespace with no value",
			cluster: clusterWithLabels(map[string]string{
				"kustomize.toolkit.fluxcd.io/name":      "test",
				"kustomize.toolkit.fluxcd.io/namespace": "",
			}),
			wantNil:       true,
			fluxNamespace: "",
			fluxName:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FluxKustomizationObjectKey(tt.cluster)
			if got != nil && tt.wantNil {
				t.Errorf("Wanted nil, got not nil")
				return
			}

			if got == nil && !tt.wantNil {
				t.Errorf("Wanted not nil, got nil")
				return
			}

			if got != nil {
				if got.Namespace != tt.fluxNamespace {
					t.Errorf("Wanted namespace to be %q, got %q", tt.fluxNamespace, got.Namespace)
					return
				}

				if got.Name != tt.fluxName {
					t.Errorf("Wanted name to be %q, got %q", tt.fluxName, got.Name)
					return
				}
			}
		})
	}
}
