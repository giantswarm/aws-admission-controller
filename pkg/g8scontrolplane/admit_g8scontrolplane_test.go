package g8scontrolplane

import (
	"github.com/ghodss/yaml"

	"github.com/giantswarm/g8s-admission-controller/pkg/admission"
)

func createG8sControlPlaneAdmitter(config []byte) (admission.Admitter, error) {
	conf := &AdmitterConfig{}

	err := yaml.Unmarshal(config, &conf)
	if err != nil {
		return nil, err
	}

	return NewAdmitter(conf)
}
