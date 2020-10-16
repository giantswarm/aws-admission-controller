package unittest

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/microerror"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CustomAdmissionRequestAWSControlPlane(AZs []string) (v1beta1.AdmissionRequest, error) {
	awsControlplane := DefaultAWSControlPlane()
	awsControlplane.Spec.AvailabilityZones = AZs
	byt, err := json.Marshal(awsControlplane)
	if err != nil {
		return v1beta1.AdmissionRequest{}, microerror.Mask(err)
	}
	req := v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AWSControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "awscontrolplanes",
		},
		Operation: v1beta1.Create,
		Object: runtime.RawExtension{
			Raw:    byt,
			Object: nil,
		},
	}
	return req, nil
}

func CustomAdmissionRequestAWSControlPlaneUpdate(oldAZs []string, newAZs []string) (v1beta1.AdmissionRequest, error) {
	// creating the old and new object for update operation
	awsControlplane := DefaultAWSControlPlane()
	awsControlplane.Spec.AvailabilityZones = oldAZs
	oldByt, err := json.Marshal(awsControlplane)
	if err != nil {
		return v1beta1.AdmissionRequest{}, microerror.Mask(err)
	}
	awsControlplane.Spec.AvailabilityZones = newAZs
	newByt, err := json.Marshal(awsControlplane)
	if err != nil {
		return v1beta1.AdmissionRequest{}, microerror.Mask(err)
	}

	req := v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AWSControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "awscontrolplanes",
		},
		Operation: v1beta1.Update,
		Object: runtime.RawExtension{
			Raw:    oldByt,
			Object: nil,
		},
		OldObject: runtime.RawExtension{
			Raw:    newByt,
			Object: nil,
		},
	}
	return req, nil
}
