package unittest

import (
	"encoding/json"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/id"
	"github.com/giantswarm/microerror"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func DefaultAdmissionRequestNetworkPool(cidrBlock string) (admissionv1.AdmissionRequest, error) {
	byt, err := json.Marshal(DefaultNetworkPool(cidrBlock))
	if err != nil {
		return admissionv1.AdmissionRequest{}, microerror.Mask(err)
	}

	req := admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Version: "infrastructure.giantswarm.io/v1alpha2",
			Kind:    "NetworkPool",
		},
		Resource: metav1.GroupVersionResource{
			Version:  "infrastructure.giantswarm.io/v1alpha2",
			Resource: "networkpools",
		},
		Operation: admissionv1.Create,
		Object: runtime.RawExtension{
			Raw:    byt,
			Object: nil,
		},
	}
	return req, nil
}

func DefaultNetworkPool(cidrBlock string) *v1alpha2.NetworkPool {
	cr := &v1alpha2.NetworkPool{
		TypeMeta: v1.TypeMeta{
			Kind:       "NetworkPool",
			APIVersion: "infrastructure.giantswarm.io/v1alpha2",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      id.Generate(),
			Namespace: id.Generate(),
			Labels: map[string]string{
				"giantswarm.io/organization": "giantswarm",
			},
		},
		Spec: v1alpha2.NetworkPoolSpec{
			CIDRBlock: cidrBlock,
		},
	}
	return cr
}
