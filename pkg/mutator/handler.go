package mutator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type Mutator interface {
	Log(keyVals ...interface{})
	Mutate(review *admissionv1.AdmissionRequest) ([]PatchOperation, error)
	Resource() string
}

var (
	scheme        = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(scheme)
	Deserializer  = codecs.UniversalDeserializer()
	InternalError = errors.New("internal admission controller error")
)

func Handler(mutator Mutator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		defer metrics.DurationRequests.WithLabelValues("mutating", mutator.Resource()).Observe(float64(time.Since(start)) / float64(time.Second))

		metrics.TotalRequests.WithLabelValues("mutating", mutator.Resource()).Inc()
		if request.Header.Get("Content-Type") != "application/json" {
			mutator.Log("level", "error", "message", fmt.Sprintf("invalid content-type: %s", request.Header.Get("Content-Type")))
			metrics.InvalidRequests.WithLabelValues("mutating", mutator.Resource()).Inc()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := io.ReadAll(request.Body)
		if err != nil {
			mutator.Log("level", "error", "message", "unable to read request")
			metrics.InternalError.WithLabelValues("mutating", mutator.Resource()).Inc()
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := admissionv1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			mutator.Log("level", "error", "message", "unable to parse admission review request")
			metrics.InvalidRequests.WithLabelValues("mutating", mutator.Resource()).Inc()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		resourceName := fmt.Sprintf("%s %s/%s", review.Request.Kind, review.Request.Namespace, handler.ExtractName(review.Request, Deserializer))

		patch, err := mutator.Mutate(review.Request)
		if err != nil {
			mutator.Log("level", "error", "message", fmt.Sprintf("error during mutation process of %s: %v", resourceName, err))
			writeResponse(mutator, writer, errorResponse(review.Request.UID, microerror.Mask(err)))
			metrics.RejectedRequests.WithLabelValues("mutating", mutator.Resource()).Inc()
			return
		}

		patchData, err := json.Marshal(patch)
		if err != nil {
			mutator.Log("level", "error", "message", fmt.Sprintf("unable to serialize patch for %s: %v", resourceName, err))
			writeResponse(mutator, writer, errorResponse(review.Request.UID, InternalError))
			metrics.RejectedRequests.WithLabelValues("mutating", mutator.Resource()).Inc()
			return
		}

		mutator.Log("level", "debug", "message", fmt.Sprintf("mutator admitted %s (with %d patches)", resourceName, len(patch)))
		metrics.SuccessfulRequests.WithLabelValues("mutating", mutator.Resource()).Inc()

		pt := admissionv1.PatchTypeJSONPatch
		writeResponse(mutator, writer, &admissionv1.AdmissionResponse{
			Allowed:   true,
			UID:       review.Request.UID,
			Patch:     patchData,
			PatchType: &pt,
		})
	}
}

func writeResponse(mutator Mutator, writer http.ResponseWriter, response *admissionv1.AdmissionResponse) {
	resp, err := json.Marshal(admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: response,
	})
	if err != nil {
		mutator.Log("level", "error", "message", "unable to serialize response", microerror.JSON(err))
		metrics.InternalError.WithLabelValues("mutating", mutator.Resource()).Inc()
		writer.WriteHeader(http.StatusInternalServerError)
	}
	if _, err := writer.Write(resp); err != nil {
		mutator.Log("level", "error", "message", "unable to write response", microerror.JSON(err))
	}
}

func errorResponse(uid types.UID, err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Allowed: false,
		UID:     uid,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
