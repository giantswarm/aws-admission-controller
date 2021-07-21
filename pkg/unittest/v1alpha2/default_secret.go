package unittest

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
)

func DefaultClusterCredentialSecretLocation() types.NamespacedName {
	return types.NamespacedName{
		Name:      "example-credential",
		Namespace: "example-namespace",
	}
}

func DefaultClusterCredentialSecret() corev1.Secret {
	cr := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.ManagedBy:    "credentiald",
				label.Organization: "example-organization",
			},
			Name:      "example-credential",
			Namespace: "example-namespace",
		},
		Type: "Opaque",
	}

	return cr
}
