package key

import (
	"strings"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
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

// Ensure the needed escapes are in place. See https://tools.ietf.org/html/rfc6901#section-3 .
func EscapeJSONPatchString(input string) string {
	input = strings.ReplaceAll(input, "~", "~0")
	input = strings.ReplaceAll(input, "/", "~1")

	return input
}
