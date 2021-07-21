package unittest

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/microerror"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func DefaultAdmissionRequestAWSCluster() (admissionv1.AdmissionRequest, error) {
	awsCluster := DefaultAWSCluster()
	byt, err := json.Marshal(awsCluster)
	if err != nil {
		return admissionv1.AdmissionRequest{}, microerror.Mask(err)
	}
	req := admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AWSCluster",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "awsclusters",
		},
		Operation: admissionv1.Create,
		Object: runtime.RawExtension{
			Raw:    byt,
			Object: nil,
		},
	}
	return req, nil
}

func DefaultAdmissionRequestAWSControlPlane() (admissionv1.AdmissionRequest, error) {
	byt, err := json.Marshal(DefaultAWSControlPlane())
	if err != nil {
		return admissionv1.AdmissionRequest{}, microerror.Mask(err)
	}

	req := admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AWSControlPlane",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "awscontrolplanes",
		},
		Operation: admissionv1.Create,
		Object: runtime.RawExtension{
			Raw:    byt,
			Object: nil,
		},
	}
	return req, nil
}

func DefaultAdmissionRequestG8sControlPlane() (admissionv1.AdmissionRequest, error) {
	byt, err := json.Marshal(DefaultG8sControlPlane())
	if err != nil {
		return admissionv1.AdmissionRequest{}, microerror.Mask(err)
	}

	req := admissionv1.AdmissionRequest{
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
			Raw:    byt,
			Object: nil,
		},
	}
	return req, nil
}

func DefaultAdmissionRequestAWSMachineDeployment() (admissionv1.AdmissionRequest, error) {
	byt, err := json.Marshal(DefaultAWSMachineDeployment())
	if err != nil {
		return admissionv1.AdmissionRequest{}, microerror.Mask(err)
	}

	req := admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "AWSMachineDeployment",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "awsmachinedeployments",
		},
		Operation: admissionv1.Create,
		Object: runtime.RawExtension{
			Raw:    byt,
			Object: nil,
		},
	}
	return req, nil
}
