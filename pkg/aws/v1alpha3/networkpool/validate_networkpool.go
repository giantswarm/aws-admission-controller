package networkpool

import (
	"context"
	"fmt"
	"net"
	"time"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/giantswarm/aws-admission-controller/v3/config"
	"github.com/giantswarm/aws-admission-controller/v3/pkg/validator"
)

type Validator struct {
	dockerCIDR               string
	ipamNetworkCIDR          string
	k8sClient                k8sclient.Interface
	kubernetesClusterIPRange string
	logger                   micrologger.Logger
}

func NewValidator(config config.Config) (*Validator, error) {
	if config.DockerCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DockerCIDR must not be empty", config)
	}
	if config.IPAMNetworkCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.IPAMNetworkCIDR must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.KubernetesClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.KubernetesClusterIPRange must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	validator := &Validator{
		dockerCIDR:               config.DockerCIDR,
		ipamNetworkCIDR:          config.IPAMNetworkCIDR,
		k8sClient:                config.K8sClient,
		kubernetesClusterIPRange: config.KubernetesClusterIPRange,
		logger:                   config.Logger,
	}

	return validator, nil
}

func (v *Validator) Validate(request *admissionv1.AdmissionRequest) (bool, error) {
	var networkPool infrastructurev1alpha3.NetworkPool
	var err error

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &networkPool); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse networkpool: %v", err)
	}
	err = v.networkPoolAllowed(networkPool)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *Validator) networkPoolAllowed(np infrastructurev1alpha3.NetworkPool) error {
	var err error
	var fetch func() error
	var networkCIDRs []string
	var networkPoolList infrastructurev1alpha3.NetworkPoolList

	// Fetch all NetworkPools.
	{
		v.Log("level", "debug", "message", "Fetching all NetworkPools")
		fetch = func() error {
			ctx := context.Background()

			err = v.k8sClient.CtrlClient().List(
				ctx,
				&networkPoolList,
			)
			if err != nil {
				return microerror.Maskf(notFoundError, "failed to fetch NetworkPools: %v", err)
			}

			return nil
		}
	}

	{
		b := backoff.NewMaxRetries(3, 10*time.Millisecond)
		err = backoff.Retry(fetch, b)
		if IsNotFound(err) {
			v.Log("level", "debug", "message", fmt.Sprintf("No NetworkPool could be found: %v", err))
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	// append all CIDRs from existing NetworkPools
	for _, networkPool := range networkPoolList.Items {
		// we do not want to add the same networkpool when updating, e.g. when extending the IP range
		if networkPool.Name == np.Name && networkPool.Namespace == np.Namespace {
			continue
		}
		networkCIDRs = append(networkCIDRs, networkPool.Spec.CIDRBlock)
	}

	// append Docker CIDR, Kubernetes cluster IP range and tenant cluster CIDR
	networkCIDRs = append(networkCIDRs, v.dockerCIDR, v.ipamNetworkCIDR, v.kubernetesClusterIPRange)

	// parse CIDRBlock from NetworkPool
	customNet, err := mustParseCIDR(np.Spec.CIDRBlock)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, cidr := range networkCIDRs {
		net, err := mustParseCIDR(cidr)
		if err != nil {
			return microerror.Mask(err)
		}
		// in case of overlapping network ranges we do not allow creating this NetworkPool
		if intersect(customNet, net) {
			return microerror.Maskf(intersectFailedError, fmt.Sprintf("network pool %s intersect with an existing CIDR %s", customNet.String(), net.String()))
		}

	}

	return nil
}

func (v *Validator) Log(keyVals ...interface{}) {
	v.logger.Log(keyVals...)
}

func (v *Validator) Resource() string {
	return "networkpool"
}

func mustParseCIDR(cidr string) (*net.IPNet, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return ipNet, nil
}

// intersect checks if network ranges overlap
func intersect(n1, n2 *net.IPNet) bool {
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}
