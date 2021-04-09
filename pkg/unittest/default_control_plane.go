package unittest

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
)

const (
	DefaultClusterID      = "8y5ck"
	DefaultControlPlaneID = "a2wax"
)

func DefaultAvailabilityZones() []string {
	return []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"}
}
func DefaultInstanceTypes() []string {
	return []string{"c5.xlarge", "c5.2xlarge", "m5.xlarge", "m4.xlarge"}
}

func DefaultAWSControlPlane() infrastructurev1alpha2.AWSControlPlane {
	cr := infrastructurev1alpha2.AWSControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a2wax",
			Labels: map[string]string{
				label.Cluster:            DefaultClusterID,
				label.ControlPlane:       DefaultControlPlaneID,
				label.AWSOperatorVersion: "7.3.0",
				label.Release:            "100.0.0",
				label.Organization:       "giantswarm",
			},
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.AWSControlPlaneSpec{
			AvailabilityZones: []string{"eu-central-1b"},
			InstanceType:      "m5.xlarge",
		},
	}

	return cr
}

func DefaultG8sControlPlane() infrastructurev1alpha2.G8sControlPlane {
	cr := infrastructurev1alpha2.G8sControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a2wax",
			Labels: map[string]string{
				label.Cluster:                DefaultClusterID,
				label.ControlPlane:           DefaultControlPlaneID,
				label.ClusterOperatorVersion: "7.3.0",
				label.Release:                "100.0.0",
				label.Organization:           "giantswarm",
			},
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.G8sControlPlaneSpec{
			Replicas: 1,
		},
	}

	return cr
}
