package g8scontrolplane

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/pkg/unittest"
)

func TestInfraRefG8sControlPlaneAdmit(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context

		expectedReferenceName string
		awsControlPlaneExists bool
	}{
		{
			// An AWSControlPlane exists
			name: "case 0",
			ctx:  context.Background(),

			awsControlPlaneExists: true,
			expectedReferenceName: controlPlaneName,
		},
		{
			// No AWSControlPlane exists
			name: "case 1",
			ctx:  context.Background(),

			awsControlPlaneExists: false,
			expectedReferenceName: "",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error
			var updatedInfraRefName string

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
			// create AWSControlPlane if needed
			if tc.awsControlPlaneExists {
				err = fakeK8sClient.CtrlClient().Create(tc.ctx, awsControlPlane([]string{"eu-central-1a", "eu-central-1b", "eu-central-1c"}))
				if err != nil {
					t.Fatal(err)
				}
			}

			// run admission request for g8sControlPlane without reference
			var patch []mutator.PatchOperation
			request, err := g8sControlPlaneNoReferenceAdmissionRequest()
			if err != nil {
				t.Fatal(err)
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

func g8sControlPlaneNoReferenceAdmissionRequest() (*v1beta1.AdmissionRequest, error) {
	g8scontrolplane, err := getG8sControlPlaneNoRefRAWByte()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	req := &v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "G8sControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "g8scontrolplanes",
		},
		Operation: v1beta1.Create,
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
				"giantswarm.io/control-plane":   controlPlaneName,
				"giantswarm.io/organization":    "giantswarm",
				"release.giantswarm.io/version": "11.5.0",
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
