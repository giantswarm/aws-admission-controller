package key

import (
	"github.com/giantswarm/aws-admission-controller/v3/pkg/label"
)

func AWSOperator(getter LabelsGetter) string {
	return getter.GetLabels()[label.AWSOperatorVersion]
}

func Cluster(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func ClusterOperator(getter LabelsGetter) string {
	return getter.GetLabels()[label.ClusterOperatorVersion]
}

func ControlPlane(getter LabelsGetter) string {
	return getter.GetLabels()[label.ControlPlane]
}
func Release(getter LabelsGetter) string {
	return getter.GetLabels()[label.Release]
}

func MachineDeployment(getter LabelsGetter) string {
	return getter.GetLabels()[label.MachineDeployment]
}

func Organization(getter LabelsGetter) string {
	return getter.GetLabels()[label.Organization]
}
