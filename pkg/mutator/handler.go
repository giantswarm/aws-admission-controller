package mutator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/giantswarm/microerror"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-admission-controller/pkg/metrics"
)

type Mutator interface {
	Log(keyVals ...interface{})
	Mutate(review *v1beta1.AdmissionRequest) ([]PatchOperation, error)
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

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			mutator.Log("level", "error", "message", "unable to read request")
			metrics.InternalError.WithLabelValues("mutating", mutator.Resource()).Inc()
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			mutator.Log("level", "error", "message", "unable to parse admission review request")
			metrics.InvalidRequests.WithLabelValues("mutating", mutator.Resource()).Inc()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		resourceName := fmt.Sprintf("%s %s/%s", review.Request.Kind, review.Request.Namespace, extractName(review.Request))

		patch, err := mutator.Mutate(review.Request)
		if err != nil {
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

		mutator.Log("level", "debug", "message", fmt.Sprintf("admitted %s (with %d patches)", resourceName, len(patch)))
		metrics.SuccessfulRequests.WithLabelValues("mutating", mutator.Resource()).Inc()

		pt := v1beta1.PatchTypeJSONPatch
		writeResponse(mutator, writer, &v1beta1.AdmissionResponse{
			Allowed:   true,
			UID:       review.Request.UID,
			Patch:     patchData,
			PatchType: &pt,
		})
	}
}

func extractName(request *v1beta1.AdmissionRequest) string {
	if request.Name != "" {
		return request.Name
	}

	obj := metav1beta1.PartialObjectMetadata{}
	if _, _, err := Deserializer.Decode(request.Object.Raw, nil, &obj); err != nil {
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

func writeResponse(mutator Mutator, writer http.ResponseWriter, response *v1beta1.AdmissionResponse) {
	resp, err := json.Marshal(v1beta1.AdmissionReview{
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

func errorResponse(uid types.UID, err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		UID:     uid,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
