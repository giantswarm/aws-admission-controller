package handler

import (
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func ExtractName(request *admissionv1.AdmissionRequest, deserializer runtime.Decoder) string {
	if request.Name != "" {
		return request.Name
	}

	obj := metav1.PartialObjectMetadata{}
	if _, _, err := deserializer.Decode(request.Object.Raw, nil, &obj); err != nil {
		return "<unknown>"
	}

	if obj.Name != "" {
		return obj.Name
	}
	if obj.GenerateName != "" {
		return obj.GenerateName + "<generated>"
	}
	return "<unknown>"
}
