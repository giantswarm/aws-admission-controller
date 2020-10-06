package networkpool

import (
	"context"
	"fmt"
	"net"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/admission/v1beta1"

	"github.com/giantswarm/aws-admission-controller/config"
	"github.com/giantswarm/aws-admission-controller/pkg/validator"
)

type Validator struct {
	ipamNetworkCIDR string
	k8sClient       k8sclient.Interface
	logger          micrologger.Logger
}

func NewValidator(config config.Config) (*Validator, error) {
	if config.IPAMNetworkCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.IPAMNetworkCIDR must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	validator := &Validator{
		ipamNetworkCIDR: config.IPAMNetworkCIDR,
		k8sClient:       config.K8sClient,
		logger:          config.Logger,
	}

	return validator, nil
}

func (v *Validator) Validate(request *v1beta1.AdmissionRequest) (bool, error) {
	var networkPool infrastructurev1alpha2.NetworkPool
	var allowed bool

	if _, _, err := validator.Deserializer.Decode(request.Object.Raw, nil, &networkPool); err != nil {
		return false, microerror.Maskf(parsingFailedError, "unable to parse networkpool: %v", err)
	}
	allowed, err := v.networkPoolOverlapping(networkPool)
	if err != nil {
		return false, microerror.Mask(err)

	}

	return allowed, nil
}

func (v *Validator) networkPoolOverlapping(networkPool infrastructurev1alpha2.NetworkPool) (bool, error) {
	var err error
	var fetch func() error
	var networkCIDRs []string
	var networkPoolList infrastructurev1alpha2.NetworkPoolList

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
		b := backoff.NewMaxRetries(3, 1*time.Second)
		err = backoff.Retry(fetch, b)
		if IsNotFound(err) {
			v.Log("level", "debug", "message", fmt.Sprintf("No NetworkPool could be found: %v", err))
		} else if err != nil {
			return false, microerror.Mask(err)
		}
	}

	// append all CIDRs from existing NetworkPools
	for _, networkPool := range networkPoolList.Items {
		networkCIDRs = append(networkCIDRs, networkPool.Spec.CIDRBlock)
	}

	// append tenant network CIDR
	networkCIDRs = append(networkCIDRs, v.ipamNetworkCIDR)

	// parse CIDRBlock from NetworkPool
	customNet, err := mustParseCIDR(networkPool.Spec.CIDRBlock)
	if err != nil {
		return false, microerror.Mask(err)
	}

	for _, cidr := range networkCIDRs {
		net, err := mustParseCIDR(cidr)
		if err != nil {
			return false, microerror.Mask(err)
		}
		// in case of overlapping network ranges we do not allow creating this NetworkPool
		if intersect(customNet, net) {
			return false, microerror.Maskf(intersectFailedError, fmt.Sprintf("network pool %s intersect with %s", customNet.String(), net.String()))
		}

	}

	return true, nil
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
