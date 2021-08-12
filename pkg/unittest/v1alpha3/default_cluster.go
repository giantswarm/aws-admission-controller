package unittest

import (
	"time"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
)

const (
	DefaultAWSOperatorVersion     = "7.3.0"
	DefaultPodCIDR                = "10.2.0.0/16"
	DefaultClusterDNSDomain       = "gauss.eu-west-1.aws.gigantic.io"
	DefaultClusterOperatorVersion = "1.1.1"
	DefaultClusterRegion          = "eu-west-1"
	DefaultReleaseVersion         = "100.0.0"
	DefaultMasterInstanceType     = "m4.xlarge"
	DefaultMasterAvailabilityZone = "eu-central-1b"
	DefaultOrganizationName       = "test-organization"
	DefaultProviderTagKey         = "tag.provider.giantswarm.io/TaggingVersion"
	DefaultProviderTagValue       = "2.4"
)

func DefaultAWSCluster() infrastructurev1alpha3.AWSCluster {
	cr := infrastructurev1alpha3.AWSCluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:            DefaultClusterID,
				label.AWSOperatorVersion: "7.3.0",
				label.Release:            "100.0.0",
				label.Organization:       "example-organization",
			},
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha3.AWSClusterSpec{
			Cluster: infrastructurev1alpha3.AWSClusterSpecCluster{
				Description: "Dev cluster",
				DNS: infrastructurev1alpha3.AWSClusterSpecClusterDNS{
					Domain: "g8s.example.com",
				},
				KubeProxy: infrastructurev1alpha3.AWSClusterSpecClusterKubeProxy{
					ConntrackMaxPerCore: 100000,
				},
				OIDC: infrastructurev1alpha3.AWSClusterSpecClusterOIDC{
					Claims: infrastructurev1alpha3.AWSClusterSpecClusterOIDCClaims{
						Username: "username-field",
						Groups:   "groups-field",
					},
					ClientID:  "some-example-client-id",
					IssuerURL: "https://idp.example.com/",
				},
			},
			Provider: infrastructurev1alpha3.AWSClusterSpecProvider{
				CredentialSecret: infrastructurev1alpha3.AWSClusterSpecProviderCredentialSecret{
					Name:      "example-credential",
					Namespace: "example-namespace",
				},
				Master: infrastructurev1alpha3.AWSClusterSpecProviderMaster{
					AvailabilityZone: "eu-central-1b",
					InstanceType:     "m5.xlarge",
				},
				Region: "eu-central-1",
			},
		},
		Status: infrastructurev1alpha3.AWSClusterStatus{
			Cluster: infrastructurev1alpha3.CommonClusterStatus{
				Conditions: []infrastructurev1alpha3.CommonClusterStatusCondition{
					{
						LastTransitionTime: metav1.Date(2020, 4, 16, 12, 51, 33, 432, time.UTC),
						Condition:          "Created",
					},
					{
						LastTransitionTime: metav1.Date(2020, 4, 16, 12, 35, 33, 432, time.UTC),
						Condition:          "Creating",
					},
				},
				ID: DefaultClusterID,
				Versions: []infrastructurev1alpha3.CommonClusterStatusVersion{
					{
						LastTransitionTime: metav1.Now(),
						Version:            "8.2.3",
					},
				},
			},
			Provider: infrastructurev1alpha3.AWSClusterStatusProvider{
				Network: infrastructurev1alpha3.AWSClusterStatusProviderNetwork{
					CIDR: "172.19.73.0/24",
				},
			},
		},
	}

	return cr
}

func DefaultCluster() *capiv1alpha3.Cluster {
	cluster := &capiv1alpha3.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
			Labels:    DefaultLabels(),
		},
		Spec: capiv1alpha3.ClusterSpec{
			InfrastructureRef: &v1.ObjectReference{
				Kind:       "AWSCluster",
				Name:       DefaultMachineDeploymentID,
				APIVersion: "infrastructure.giantswarm.io/v1alpha2",
			},
		},
	}

	return cluster
}

func DefaultClusterEmptyOrganization() *capiv1alpha3.Cluster {
	cluster := &capiv1alpha3.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				label.Cluster:                DefaultClusterID,
				label.ClusterOperatorVersion: "1.2.3",
				label.Release:                "100.0.0",
				label.Organization:           "",
			},
		},
	}

	return cluster
}

func DefaultClusterUnknownOrganization() *capiv1alpha3.Cluster {
	cluster := &capiv1alpha3.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				label.Cluster:                DefaultClusterID,
				label.ClusterOperatorVersion: "1.2.3",
				label.Release:                "100.0.0",
				label.Organization:           "unknown-organization",
			},
		},
	}

	return cluster
}
func DefaultClusterWithoutOrganizationLabel() *capiv1alpha3.Cluster {
	cluster := &capiv1alpha3.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				label.Cluster:                DefaultClusterID,
				label.ClusterOperatorVersion: "1.2.3",
				label.Release:                "100.0.0",
			},
		},
	}

	return cluster
}
