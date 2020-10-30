// Package awsmachinedeployment intercepts write activity to AWSMachineDeployment objects.
package awscluster

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
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

	mutator := &Mutator{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		podCIDRBlock: fmt.Sprintf("%s/%s", config.PodSubnet, config.PodCIDR),
	}

	return mutator, nil
}

// Mutate is the function executed for every matching webhook request.
func (m *Mutator) Mutate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation

	if request.DryRun != nil && *request.DryRun {
		return result, nil
	}
	if request.Operation == admissionv1.Create {
		return m.MutateCreate(request)
	}
	if request.Operation == admissionv1.Update {
		return m.MutateUpdate(request)
	}
	return result, nil
}

// MutateCreate is the function executed for every create webhook request.
func (m *Mutator) MutateCreate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	awsCluster := &infrastructurev1alpha2.AWSCluster{}
	if _, _, err = mutator.Deserializer.Decode(request.Object.Raw, nil, awsCluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse AWSCluster: %v", err)
	}
	patch, err = m.MutatePodCIDR(*awsCluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

// MutateUpdate is the function executed for every update webhook request.
func (m *Mutator) MutateUpdate(request *admissionv1.AdmissionRequest) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	var patch []mutator.PatchOperation
	var err error

	awsCluster := &infrastructurev1alpha2.AWSCluster{}
	if _, _, err = mutator.Deserializer.Decode(request.Object.Raw, nil, awsCluster); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse AWSCluster: %v", err)
	}
	awsClusterOld := &infrastructurev1alpha2.AWSCluster{}
	if _, _, err = mutator.Deserializer.Decode(request.OldObject.Raw, nil, awsClusterOld); err != nil {
		return nil, microerror.Maskf(parsingFailedError, "unable to parse old AWSCluster: %v", err)
	}
	patch, err = m.MutatePodCIDR(*awsCluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	result = append(result, patch...)

	return result, nil
}

// MutatePodCIDR defaults the Pod CIDR if it is not set.
func (m *Mutator) MutatePodCIDR(awsCluster infrastructurev1alpha2.AWSCluster) ([]mutator.PatchOperation, error) {
	var result []mutator.PatchOperation
	if &awsCluster.Spec.Provider.Pods != nil {
		if awsCluster.Spec.Provider.Pods.CIDRBlock != "" {
			return result, nil
		}
		if awsCluster.Spec.Provider.Pods.ExternalSNAT != nil {
			// If the Pod CIDR is not set but the pods attribute exists, we default here
			m.Log("level", "debug", "message", fmt.Sprintf("AWSCluster %s Pod CIDR Block is not set and will be defaulted to %s",
				awsCluster.ObjectMeta.Name,
				m.podCIDRBlock),
			)
			patch := mutator.PatchAdd("/spec/provider/pods/", "cidrBlock")
			result = append(result, patch)
			patch = mutator.PatchAdd("/spec/provider/pods/cidrBlock", m.podCIDRBlock)
			result = append(result, patch)
			return result, nil
		}
	}
	// If the Pod CIDR is not set we default it here
	m.Log("level", "debug", "message", fmt.Sprintf("AWSCluster %s Pod CIDR Block is not set and will be defaulted to %s",
		awsCluster.ObjectMeta.Name,
		m.podCIDRBlock),
	)
	patch := mutator.PatchAdd("/spec/provider/", "pods")
	result = append(result, patch)
	patch = mutator.PatchAdd("/spec/provider/pods", map[string]string{"cidrBlock": m.podCIDRBlock})
	result = append(result, patch)

	return result, nil
}

func (m *Mutator) Log(keyVals ...interface{}) {
	m.logger.Log(keyVals...)
}

func (m *Mutator) Resource() string {
	return "awscluster"
}
