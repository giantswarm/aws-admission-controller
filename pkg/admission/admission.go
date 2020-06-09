package admission

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
)

type Admitter interface {
	Admit(review *v1beta1.AdmissionRequest) ([]PatchOperation, error)
}

var (
	scheme        = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(scheme)
	Deserializer  = codecs.UniversalDeserializer()
	InternalError = errors.New("internal admission controller error")
)

func Handler(admitter Admitter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/json" {
			log.Errorf("Invalid content-type: %s", request.Header.Get("Content-Type"))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Errorf("Unable to read request: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		review := v1beta1.AdmissionReview{}
		if _, _, err := Deserializer.Decode(data, nil, &review); err != nil {
			log.Errorf("Unable to parse request: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		} else {
			resourceName := fmt.Sprintf("%s %s/%s", review.Request.Kind, review.Request.Namespace, extractName(review.Request))

			patch, err := admitter.Admit(review.Request)
			if err != nil {
				writeResponse(writer, errorResponse(review.Request.UID, err))
				return
			}

			patchData, err := json.Marshal(patch)
			if err != nil {
				log.Errorf("Unable to serialize patch for %s: %v", resourceName, err)
				writeResponse(writer, errorResponse(review.Request.UID, InternalError))
				return
			}

			log.Infof("Admitted %s (with %d patches)", resourceName, len(patch))

			pt := v1beta1.PatchTypeJSONPatch
			writeResponse(writer, &v1beta1.AdmissionResponse{
				Allowed:   true,
				UID:       review.Request.UID,
				Patch:     patchData,
				PatchType: &pt,
			})
		}
	}
}

func extractName(request *v1beta1.AdmissionRequest) string {
	if request.Name != "" {
		return request.Name
	}

	obj := metav1beta1.PartialObjectMetadata{}
	if _, _, err := Deserializer.Decode(request.Object.Raw, nil, &obj); err != nil {
		log.Warnf("unable to parse object: %v", err)
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

func writeResponse(writer http.ResponseWriter, response *v1beta1.AdmissionResponse) {
	resp, err := json.Marshal(v1beta1.AdmissionReview{
		Response: response,
	})
	if err != nil {
		log.Errorf("unable to serialize response: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
	if _, err := writer.Write(resp); err != nil {
		log.Errorf("unable to write response: %v", err)
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
