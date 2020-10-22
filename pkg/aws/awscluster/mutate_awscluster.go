// Package awsmachinedeployment intercepts write activity to AWSMachineDeployment objects.
package awscluster

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/config"
	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

// Mutator for AWSMachineDeployment object.
type Mutator struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	podCIDRBlock string
}

func NewMutator(config config.Config) (*Mutator, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var cidrBlock string = fmt.Sprintf("%s/%s", config.PodSubnet, config.PodCIDR)
	mutator := &Mutator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		podCIDRBlock: cidrBlock,
	}

	return mutator, nil
}

// Mutate is the function executed for every matching webhook request.
func (m *Mutator) Mutate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		return result, nil
	}
	awsCluster := &infrastructurev1alpha2.AWSCluster{}
	if _, _, err := mutator.Deserializer.Decode(request.Object.Raw, nil, awsCluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse AWSCluster: %v", err)
	}
	if &awsCluster.Spec.Provider.Pods != nil {
		if awsCluster.Spec.Provider.Pods.CIDRBlock != "" {
			return result, nil
		}
	}
	// If the Pod CIDR is not set we default it here
	m.Log("level", "debug", "message", fmt.Sprintf("AWSCluster %s Pod CIDR Block is not set and will be defaulted to %s",
		awsCluster.ObjectMeta.Name,
		m.podCIDRBlock),
	)
	patch := mutator.PatchAdd("/spec/provider/pods/cidrBlock", m.podCIDRBlock)
	result = append(result, patch)

	return result, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "awsmachinedeployment"
}
