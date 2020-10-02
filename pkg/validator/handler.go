package validator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/giantswarm/microerror"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/pkg/metrics"
)

type Validator interface {
	Log(keyVals ...interface{})
	Resource() string
	Validate(review *v1beta1.AdmissionRequest) (bool, error)
}

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	Deserializer = codecs.UniversalDeserializer()
)

func Handler(validator Validator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		defer metrics.DurationRequests.WithLabelValues("validating", validator.Resource()).Observe(float64(time.Since(start)) / float64(time.Second))

		if request.Header.Get("Content-Type") != "application/json" {
			validator.Log("level", "error", "message", fmt.Sprintf("invalid content-type: %s", request.Header.Get("Content-Type")))
			metrics.InvalidRequests.WithLabelValues("validating", validator.Resource()).Inc()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			validator.Log("level", "error", "message", "unable to read request")
			metrics.InternalError.WithLabelValues("validating", validator.Resource()).Inc()
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			validator.Log("level", "error", "message", "unable to parse admission review request")
			metrics.InvalidRequests.WithLabelValues("validating", validator.Resource()).Inc()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		allowed, err := validator.Validate(review.Request)
		if err != nil {
			writeResponse(validator, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
			metrics.RejectedRequests.WithLabelValues("validating", validator.Resource()).Inc()
			return
		}

		metrics.SuccessfulRequests.WithLabelValues("validating", validator.Resource()).Inc()

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
		metrics.InternalError.WithLabelValues("validating", validator.Resource()).Inc()
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
