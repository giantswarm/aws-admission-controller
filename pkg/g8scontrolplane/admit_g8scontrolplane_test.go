package g8scontrolplane

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/giantswarm/g8s-admission-controller/pkg/admission"
	"github.com/giantswarm/g8s-admission-controller/pkg/testrunner"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func createG8sControlPlaneAdmitter(config []byte) (admission.Admitter, error) {
	conf := &AdmitterConfig{}

	err := yaml.Unmarshal(config, &conf)
	if err != nil {
		return nil, err
	}

	return NewAdmitter(conf)
}

func TestJobAdmission(t *testing.T) {
	runner := &testrunner.Runner{
		CreateAdmitter: createG8sControlPlaneAdmitter,
		Resource:       g8sControlPlaneResource,
		NewElement:     func() runtime.Object { return &corev1.Node{} },
	}
	runner.RunTestcases(t)
}
