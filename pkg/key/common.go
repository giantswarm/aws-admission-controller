package key

import (
	"strconv"

	"github.com/dylanmei/iso8601"

	"github.com/giantswarm/aws-admission-controller/v2/pkg/label"
)

const (
	// annotations should  taken from https://github.com/giantswarm/apiextensions/blob/master/pkg/annotation/aws.go
	// once the service is migrate to apiextensions v3
	AnnotationUpdateMaxBatchSize = "alpha.aws.giantswarm.io/update-max-batch-size"
	AnnotationUpdatePauseTime    = "alpha.aws.giantswarm.io/update-pause-time"
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

// MaxBatchSizeIsValid will validate the value into valid maxBatchSize
// valid values can be either:
// an integer between 0 < x <= worker count
// a float between 0 < x <= 1
// float value is used as ratio of a total worker count
func MaxBatchSizeIsValid(value string, maxWorkers int) bool {
	// try parse an integer
	integer, err := strconv.Atoi(value)
	if err == nil {
		// check if the value is bigger than zero but lower-or-equal to maximum number of workers
		if integer > 0 && integer <= maxWorkers {
			// integer value can be directly used, no need for any adjustment
			return true
		} else {
			// the value is outside of valid bounds, it cannot be used
			return false
		}
	}
	// try parse float
	ratio, err := strconv.ParseFloat(value, 10)
	if err != nil {
		// not integer or float which means invalid value
		return false
	}
	// valid value is a decimal representing a percentage
	// anything smaller than 0 or bigger than 1 is not valid
	if ratio > 0 && ratio <= 1.0 {
		return true
	}

	return false
}

// PauseTimeIsValid checks if the value is in proper ISO 8601 duration format
func PauseTimeIsValid(value string) bool {

	_, err := iso8601.ParseDuration(value)
	return err == nil
}
