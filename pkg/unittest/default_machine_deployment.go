package unittest

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-admission-controller/pkg/label"
)

const (
	DefaultMachineDeploymentID = "al9qy"
)

func DefaultMachineDeployment() apiv1alpha2.MachineDeployment {
	cr := apiv1alpha2.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: "cluster.x-k8s.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultMachineDeploymentID,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				label.Cluster:           DefaultClusterID,
				label.MachineDeployment: DefaultMachineDeploymentID,
				label.OperatorVersion:   "7.3.0",
				label.Release:           "100.0.0",
			},
		},
		Status: apiv1alpha2.MachineDeploymentStatus{
			ObservedGeneration:  0,
			Selector:            "",
			Replicas:            1,
			UpdatedReplicas:     2,
			ReadyReplicas:       1,
			AvailableReplicas:   1,
			UnavailableReplicas: 0,
		},
	}
	return cr
}

func DefaultAWSMachineDeployment() infrastructurev1alpha2.AWSMachineDeployment {
	cr := infrastructurev1alpha2.AWSMachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:           DefaultClusterID,
				label.MachineDeployment: DefaultMachineDeploymentID,
				label.OperatorVersion:   "7.3.0",
				label.Release:           "100.0.0",
			},
			Name:      DefaultMachineDeploymentID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha2.AWSMachineDeploymentSpec{
			NodePool: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePool{
				Description: "Test node pool for cluster in template rendering unit test.",
				Machine: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolMachine{
					DockerVolumeSizeGB:  100,
					KubeletVolumeSizeGB: 100,
				},
				Scaling: infrastructurev1alpha2.AWSMachineDeploymentSpecNodePoolScaling{
					Max: 5,
					Min: 3,
				},
			},
			Provider: infrastructurev1alpha2.AWSMachineDeploymentSpecProvider{
				AvailabilityZones: []string{"eu-central-1a", "eu-central-1c"},
				InstanceDistribution: infrastructurev1alpha2.AWSMachineDeploymentSpecInstanceDistribution{
					OnDemandBaseCapacity:                0,
					OnDemandPercentageAboveBaseCapacity: to.IntP(100),
				},
				Worker: infrastructurev1alpha2.AWSMachineDeploymentSpecProviderWorker{
					InstanceType:          "m5.2xlarge",
					UseAlikeInstanceTypes: true,
				},
			},
		},
	}

	return cr
}
