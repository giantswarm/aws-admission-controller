package validator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/giantswarm/microerror"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
)

type Validator interface {
	Validate(review *v1beta1.AdmissionRequest) (bool, error)
	Log(keyVals ...interface{})
}

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	Deserializer = codecs.UniversalDeserializer()
)

func Handler(validator Validator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/json" {
			validator.Log("level", "error", "message", fmt.Sprintf("invalid content-type: %s", request.Header.Get("Content-Type")))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			validator.Log("level", "error", "message", "unable to read request")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			validator.Log("level", "error", "message", "unable to parse admission review request")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		allowed, err := validator.Validate(review.Request)
		if err != nil {
			writeResponse(validator, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
			return
		}

		writeResponse(validator, writer, &v1beta1.AdmissionResponse{
			Allowed: allowed,
			UID:     review.Request.UID,
		})
	}
}

func writeResponse(validator Validator, writer http.ResponseWriter, response *v1beta1.AdmissionResponse) {
	resp, err := json.Marshal(v1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	})
	if err != nil {
		validator.Log("level", "error", "message", "unable to serialize response", microerror.JSON(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}

	if _, err := writer.Write(resp); err != nil {
		validator.Log("level", "error", "message", "unable to write response", microerror.JSON(err))
	}
}

func errorResponse(uid types.UID, err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		UID:     uid,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
