package unittest

import (
	"encoding/json"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
)

func CustomAdmissionRequestAWSCluster(podCIDRBlock string) (admissionv1.AdmissionRequest, error) {
	awsCluster := DefaultAWSCluster()
	awsCluster.Spec.Provider.Pods = infrastructurev1alpha2.AWSClusterSpecProviderPods{
		CIDRBlock: podCIDRBlock,
	}
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

func CustomAdmissionRequestAWSControlPlane(AZs []string) (admissionv1.AdmissionRequest, error) {
	awsControlplane := DefaultAWSControlPlane()
	awsControlplane.Spec.AvailabilityZones = AZs
	byt, err := json.Marshal(awsControlplane)
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

func CustomAdmissionRequestAWSControlPlaneUpdate(oldAZs []string, newAZs []string) (admissionv1.AdmissionRequest, error) {
	// creating the old and new object for update operation
	awsControlplane := DefaultAWSControlPlane()
	awsControlplane.Spec.AvailabilityZones = oldAZs
	oldByt, err := json.Marshal(awsControlplane)
	if err != nil {
		return admissionv1.AdmissionRequest{}, microerror.Mask(err)
	}
	awsControlplane.Spec.AvailabilityZones = newAZs
	newByt, err := json.Marshal(awsControlplane)
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
		Operation: admissionv1.Update,
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
