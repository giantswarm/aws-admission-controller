package awsmachinedeployment

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-admission-controller/v3/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v3/pkg/unittest/v1alpha3"
)

var (
	clusterID                  = "myc1"
	machineDeploymentID        = "mymd1"
	machineDeploymentNamespace = "mymd1"
)

func TestAWSMachineDeploymentAdmit(t *testing.T) {
	testCases := []struct {
		name                    string
		ctx                     context.Context
		currentAvailabilityZone []string
		// expectAvailabilityZones needs to be in order
		expectAvailabilityZones []string
		validAvailabilityZones  []string
	}{
		{
			name: "case 0",
			ctx:  context.Background(),
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error

			fakeK8sClient := unittest.FakeK8sClient()
			mutator := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}

			// create AWSMachineDeployment
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsMachineDeployment())
			if err != nil {
				t.Fatal(err)
			}

			// run admission request to update AWSControlPlane AZ's
			request, err := awsMachineDeploymentAdmissionRequest()
			if err != nil {
				t.Fatal(err)
			}
			_, err = mutator.Mutate(request)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAWSMachineDeploymentAvailabilityZones(t *testing.T) {
	testCases := []struct {
		ctx  context.Context
		name string

		currentAZ     []string
		expectedPatch []string
	}{
		{
			// Don't default the AZ if they are set
			name: "case 0",
			ctx:  context.Background(),

			currentAZ:     unittest.DefaultAvailabilityZones(),
			expectedPatch: nil,
		},
		{
			// Default the AZ they are not set
			name: "case 1",
			ctx:  context.Background(),

			currentAZ:     nil,
			expectedPatch: []string{unittest.DefaultMasterAvailabilityZone},
		},
	}
	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedAZs []string

			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				k8sClient: fakeK8sClient,
				logger:    microloggertest.New(),
			}
			awsControlPlane := unittest.DefaultAWSControlPlane()
			err = fakeK8sClient.CtrlClient().Create(tc.ctx, &awsControlPlane)
			if err != nil {
				t.Fatal(err)
			}

			// run mutate function to default AWSMachineDeployment AZs
			var patch []mutator.PatchOperation
			awsmachinedeployment := unittest.DefaultAWSMachineDeployment()
			awsmachinedeployment.Spec.Provider.AvailabilityZones = tc.currentAZ
			patch, err = mutate.MutateAvailabilityZones(*awsmachinedeployment)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/provider/availabilityZones" {
					updatedAZs = p.Value.([]string)
				}
			}

			// check if the AZs are patched as expected
			if len(tc.expectedPatch) != len(updatedAZs) {
				t.Fatalf("expected %v to not to differ from %v", len(tc.expectedPatch), len(updatedAZs))
			}
			for i, p := range updatedAZs {
				if tc.expectedPatch[i] != p {
					t.Fatalf("expected %v to not to differ from %v", tc.expectedPatch[i], p)
				}
			}
		})
	}
}

func awsMachineDeploymentAdmissionRequest() (*admissionv1.AdmissionRequest, error) {
	awsmachinedeployment, err := awsMachineDeploymentRawByte()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	req := &admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha3",
			Kind:    "AWSMachineDeployment",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha3",
			Resource: "awsmachinedeployments",
		},
		Operation: admissionv1.Update,
		Object: runtime.RawExtension{
			Raw:    awsmachinedeployment,
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    awsmachinedeployment,
			Object: nil,
		},
	}
	return req, nil
}

func awsMachineDeploymentRawByte() ([]byte, error) {
	var ten int = 10
	cr := infrastructurev1alpha3.AWSMachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSMachineDeployment",
			APIVersion: "infrastructure.giantswarm.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineDeploymentID,
			Namespace: machineDeploymentNamespace,
			Labels: map[string]string{
				"giantswarm.io/cluster":         clusterID,
				"giantswarm.io/control-plane":   machineDeploymentID,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": "11.5.0",
			},
		},
		Spec: infrastructurev1alpha3.AWSMachineDeploymentSpec{
			NodePool: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePool{
				Description: "Some friendly name",
				Scaling: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePoolScaling{
					Min: 1,
					Max: 30,
				},
			},
			Provider: infrastructurev1alpha3.AWSMachineDeploymentSpecProvider{
				InstanceDistribution: infrastructurev1alpha3.AWSMachineDeploymentSpecInstanceDistribution{
					OnDemandBaseCapacity:                10,
					OnDemandPercentageAboveBaseCapacity: &ten,
				},
			},
		},
	}
	byt, err := json.Marshal(cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return byt, nil
}

func awsMachineDeployment() *infrastructurev1alpha3.AWSMachineDeployment {
	var ten int = 10
	return &infrastructurev1alpha3.AWSMachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSMachineDeployment",
			APIVersion: "infrastructure.giantswarm.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineDeploymentID,
			Namespace: machineDeploymentNamespace,
			Labels: map[string]string{
				"giantswarm.io/cluster":         clusterID,
				"giantswarm.io/control-plane":   machineDeploymentID,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": "11.5.0",
			},
		},
		Spec: infrastructurev1alpha3.AWSMachineDeploymentSpec{
			NodePool: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePool{
				Description: "Some friendly name",
				Scaling: infrastructurev1alpha3.AWSMachineDeploymentSpecNodePoolScaling{
					Min: 1,
					Max: 30,
				},
			},
			Provider: infrastructurev1alpha3.AWSMachineDeploymentSpecProvider{
				InstanceDistribution: infrastructurev1alpha3.AWSMachineDeploymentSpecInstanceDistribution{
					OnDemandBaseCapacity:                10,
					OnDemandPercentageAboveBaseCapacity: &ten,
				},
			},
		},
	}
}
