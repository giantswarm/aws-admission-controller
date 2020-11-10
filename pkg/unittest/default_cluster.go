package unittest

import (
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
)

const (
	DefaultPodCIDR          = "10.2.0.0/16"
	DefaultClusterDNSDomain = "gauss.eu-west-1.aws.gigantic.io"
	DefaultClusterRegion    = "eu-west-1"
	DefaultReleaseVersion   = "100.0.0"
)

func DefaultAWSCluster() infrastructurev1alpha2.AWSCluster {
	cr := infrastructurev1alpha2.AWSCluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.OperatorVersion: "7.3.0",
				label.Release:         "100.0.0",
				label.Organization:    "example-organization",
			},
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.AWSClusterSpec{
			Cluster: infrastructurev1alpha2.AWSClusterSpecCluster{
				Description: "Dev cluster",
				DNS: infrastructurev1alpha2.AWSClusterSpecClusterDNS{
					Domain: "g8s.example.com",
				},
				KubeProxy: infrastructurev1alpha2.AWSClusterSpecClusterKubeProxy{
					ConntrackMaxPerCore: 100000,
				},
				OIDC: infrastructurev1alpha2.AWSClusterSpecClusterOIDC{
					Claims: infrastructurev1alpha2.AWSClusterSpecClusterOIDCClaims{
						Username: "username-field",
						Groups:   "groups-field",
					},
					ClientID:  "some-example-client-id",
					IssuerURL: "https://idp.example.com/",
				},
			},
			Provider: infrastructurev1alpha2.AWSClusterSpecProvider{
				CredentialSecret: infrastructurev1alpha2.AWSClusterSpecProviderCredentialSecret{
					Name:      "example-credential",
					Namespace: "example-namespace",
				},
				Master: infrastructurev1alpha2.AWSClusterSpecProviderMaster{
					AvailabilityZone: "eu-central-1b",
					InstanceType:     "m5.xlarge",
				},
				Region: "eu-central-1",
			},
		},
		Status: infrastructurev1alpha2.AWSClusterStatus{
			Cluster: infrastructurev1alpha2.CommonClusterStatus{
				Conditions: []infrastructurev1alpha2.CommonClusterStatusCondition{
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
				Versions: []infrastructurev1alpha2.CommonClusterStatusVersion{
					{
						LastTransitionTime: metav1.Now(),
						Version:            "8.2.3",
					},
				},
			},
			Provider: infrastructurev1alpha2.AWSClusterStatusProvider{
				Network: infrastructurev1alpha2.AWSClusterStatusProviderNetwork{
					CIDR: "172.19.73.0/24",
				},
			},
		},
	}

	return cr
}

func DefaultCluster() capiv1alpha2.Cluster {
	cluster := capiv1alpha2.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.OperatorVersion: "1.2.3",
				label.Release:         "100.0.0",
				label.Organization:    "example-organization",
			},
		},
	}

	return cluster
}
