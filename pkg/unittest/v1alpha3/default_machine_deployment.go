package unittest

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/to"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
)

const (
	DefaultMachineDeploymentID = "al9qy"
)

func DefaultMachineDeployment() *capi.MachineDeployment {
	cr := &capi.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: "cluster.x-k8s.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultMachineDeploymentID,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				label.Cluster:                DefaultClusterID,
				label.MachineDeployment:      DefaultMachineDeploymentID,
				label.ClusterOperatorVersion: "7.3.0",
				label.Release:                "100.0.0",
			},
		},
		Spec: capi.MachineDeploymentSpec{
			Template: capi.MachineTemplateSpec{
				Spec: capi.MachineSpec{
					InfrastructureRef: v1.ObjectReference{
						Kind:       "AWSMachineDeployment",
						Name:       DefaultMachineDeploymentID,
						APIVersion: "infrastructure.giantswarm.io/v1alpha2",
					},
				},
			},
		},
		Status: capi.MachineDeploymentStatus{
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

func DefaultAWSMachineDeployment() *infrastructurev1alpha3.AWSMachineDeployment {
	cr := &infrastructurev1alpha3.AWSMachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSMachineDeployment",
			APIVersion: "infrastructure.giantswarm.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:            DefaultClusterID,
				label.MachineDeployment:  DefaultMachineDeploymentID,
				label.AWSOperatorVersion: "7.3.0",
				label.Release:            "100.0.0",
			},
			Name:      DefaultMachineDeploymentID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: infrastructurev1alpha3.AWSMachineDeploymentSpec{
			NodePool: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePool{
				Description: "Test node pool for cluster in template rendering unit test.",
				Machine: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePoolMachine{
					DockerVolumeSizeGB:  100,
					KubeletVolumeSizeGB: 100,
				},
				Scaling: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePoolScaling{
					Max: 5,
					Min: 3,
				},
			},
			Provider: infrastructurev1alpha3.AWSMachineDeploymentSpecProvider{
				AvailabilityZones: []string{"eu-central-1a", "eu-central-1c"},
				InstanceDistribution: infrastructurev1alpha3.AWSMachineDeploymentSpecInstanceDistribution{
					OnDemandBaseCapacity:                0,
					OnDemandPercentageAboveBaseCapacity: to.IntP(100),
				},
				Worker: infrastructurev1alpha3.AWSMachineDeploymentSpecProviderWorker{
					InstanceType:          "m5.2xlarge",
					UseAlikeInstanceTypes: true,
				},
			},
		},
	}

	return cr
}
