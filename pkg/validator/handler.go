package validator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/giantswarm/microerror"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/v4/pkg/handler"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/metrics"
)

type Validator interface {
	Log(keyVals ...interface{})
	Resource() string
	Validate(review *admissionv1.AdmissionRequest) (bool, error)
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

		review := admissionv1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			validator.Log("level", "error", "message", "unable to parse admission review request")
			metrics.InvalidRequests.WithLabelValues("validating", validator.Resource()).Inc()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		resourceName := fmt.Sprintf("%s %s/%s", review.Request.Kind, review.Request.Namespace, handler.ExtractName(review.Request, Deserializer))

		allowed, err := validator.Validate(review.Request)
		if err != nil {
			validator.Log("level", "error", "message", fmt.Sprintf("error during validation process of %s: %v", resourceName, err))
			writeResponse(validator, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
			metrics.RejectedRequests.WithLabelValues("validating", validator.Resource()).Inc()
			return
		}
		validator.Log("level", "debug", "message", fmt.Sprintf("validator admitted %s", resourceName))
		metrics.SuccessfulRequests.WithLabelValues("validating", validator.Resource()).Inc()

		writeResponse(validator, writer, &admissionv1.AdmissionResponse{
			Allowed: allowed,
			UID:     review.Request.UID,
		})
	}
}

func writeResponse(validator Validator, writer http.ResponseWriter, response *admissionv1.AdmissionResponse) {
	resp, err := json.Marshal(admissionv1.AdmissionReview{
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

func errorResponse(uid types.UID, err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Allowed: false,
		UID:     uid,
		Result: &metav1.Status{
			Reason:  metav1.StatusReasonBadRequest,
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		},
	}
}
