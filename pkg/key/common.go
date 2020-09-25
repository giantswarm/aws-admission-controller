package key

import (
	"github.com/giantswarm/aws-admission-controller/pkg/label"
)

func Cluster(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func ControlPlane(getter LabelsGetter) string {
	return getter.GetLabels()[label.ControlPlane]
}

func MachineDeployment(getter LabelsGetter) string {
	return getter.GetLabels()[label.MachineDeployment]
}
