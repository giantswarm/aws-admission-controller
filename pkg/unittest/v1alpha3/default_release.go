package unittest

import (
	releasev1alpha1 "github.com/giantswarm/release-operator/v4/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultReleaseName = "v100.0.0"
)

type ReleaseData struct {
	Name  string
	State releasev1alpha1.ReleaseState
}

func DefaultRelease() releasev1alpha1.Release {
	return NamedRelease(DefaultReleaseName)
}

func NamedRelease(name string) releasev1alpha1.Release {
	cr := releasev1alpha1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: releasev1alpha1.ReleaseSpec{
			Components: []releasev1alpha1.ReleaseSpecComponent{
				{
					Name:    "aws-operator",
					Version: DefaultAWSOperatorVersion,
				},
				{
					Name:    "cluster-operator",
					Version: DefaultClusterOperatorVersion,
				},
			},
		},
	}

	return cr
}
