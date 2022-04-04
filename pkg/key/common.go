package key

import (
	"fmt"

	"github.com/giantswarm/aws-admission-controller/v4/pkg/internal/normalize"
	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
)

const (
	organizationNamespaceFormat = "org-%s"
	MinEBSVolumeThroughtput     = 100
	MaxEBSVolumeThroughtput     = 1000
	MinEBSVolumeIops            = 3000
	MaxEBSVolumeIops            = 16000
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

func OrganizationNamespaceFromName(name string) string {
	name = normalize.AsDNSLabelName(fmt.Sprintf(organizationNamespaceFormat, name))

	return name
}
