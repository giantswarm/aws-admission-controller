package g8scontrolplane

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	unittest "github.com/giantswarm/aws-admission-controller/v2/pkg/unittest/v1alpha2"
)

func TestInfraRefG8sControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context

		expectedReferenceName string
		reference             bool
	}{
		{
			// Reference not set
			name: "case 0",
			ctx:  context.Background(),

			reference:             false,
			expectedReferenceName: controlPlaneName,
		},
		{
			// Reference set
			name: "case 1",
			ctx:  context.Background(),

			reference:             true,
			expectedReferenceName: "",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedInfraRefName string
			var patch []mutator.PatchOperation
			var request *admissionv1.AdmissionRequest

			// Create a new logger that is used by all admitters.
			var newLogger micrologger.Logger
			{
				newLogger, err = micrologger.New(micrologger.Config{})
				if err != nil {
					panic(microerror.JSON(err))
				}
			}
			fakeK8sClient := unittest.FakeK8sClient()
			mutate := &Mutator{
				validAvailabilityZones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"},
				k8sClient:              fakeK8sClient,
				logger:                 newLogger,
			}

			if !tc.reference {
				// run admission request for g8sControlPlane without reference
				request, err = g8sControlPlaneNoReferenceAdmissionRequest()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				// run admission request for g8sControlPlane with reference
				request, err = g8sControlPlaneCreateAdmissionRequest(1, "11.5.0")
				if err != nil {
					t.Fatal(err)
				}
			}
			patch, err = mutate.Mutate(request)
			if err != nil {
				t.Fatal(err)
			}
			// parse patches
			for _, p := range patch {
				if p.Path == "/spec/infrastructureRef" {
					updatedInfraRefName = p.Value.(*v1.ObjectReference).Name
				}
			}
			// check if the reference patch is as expected
			if tc.expectedReferenceName != updatedInfraRefName {
				t.Fatalf("expected %#q to be equal to %#q", tc.expectedReferenceName, updatedInfraRefName)
			}
		})
	}
}

func g8sControlPlaneNoReferenceAdmissionRequest() (*admissionv1.AdmissionRequest, error) {
	g8scontrolplane, err := getG8sControlPlaneNoRefRAWByte()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	req := &admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "G8sControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "g8scontrolplanes",
		},
		Operation: admissionv1.Create,
		Object: runtime.RawExtension{
			Raw:    g8scontrolplane,
			Object: nil,
		},
	}
	return req, nil
}

func getG8sControlPlaneNoRefRAWByte() ([]byte, error) {
	g8scontrolPlane := &infrastructurev1alpha2.G8sControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "G8sControlPlane",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controlPlaneName,
			Namespace: controlPlaneNameSpace,
			Labels: map[string]string{
				"giantswarm.io/control-plane":            controlPlaneName,
				"giantswarm.io/cluster":                  clusterName,
				"cluster-operator.giantswarm.io/version": "1.2.3",
				"giantswarm.io/organization":             "giantswarm",
				"release.giantswarm.io/version":          "11.5.0",
			},
		},
		Spec: infrastructurev1alpha2.G8sControlPlaneSpec{
			Replicas:          1,
			InfrastructureRef: v1.ObjectReference{},
		},
	}
	byt, err := json.Marshal(g8scontrolPlane)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return byt, nil
}
