package unittest

import (
	"github.com/giantswarm/aws-admission-controller/v4/pkg/label"
)

func DefaultLabels() map[string]string {
	return map[string]string{
		label.Cluster:                DefaultClusterID,
		label.ClusterOperatorVersion: "1.2.3",
		label.Release:                "100.0.0",
		label.Organization:           "example-organization",
		"example-key":                "example-value",
		DefaultProviderTagKey:        DefaultProviderTagValue,
		label.ServicePriority:        "highest",
	}
}
