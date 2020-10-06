package config

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	restclient "k8s.io/client-go/rest"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

const (
	defaultAddress        = ":8443"
	defaultMetricsAddress = ":8080"
)

type Config struct {
	Address           string
	MetricsAddress    string
	AvailabilityZones string
	CertFile          string
	Logger            micrologger.Logger
	K8sClient         k8sclient.Interface
	KeyFile           string
}

func Parse() (Config, error) {
	var err error
	var config Config

	// Create a new logger that is used by all admitters.
	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
		config.Logger = newLogger
	}

	// Create a new k8sclient that is used by all admitters.
	var k8sClient k8sclient.Interface
	{
		restConfig, err := restclient.InClusterConfig()
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
		c := k8sclient.ClientsConfig{
			SchemeBuilder: k8sclient.SchemeBuilder{
				apiv1alpha2.AddToScheme,
				infrastructurev1alpha2.AddToScheme,
				releasev1alpha1.AddToScheme,
			},
			Logger: config.Logger,

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
		config.K8sClient = k8sClient
	}

	kingpin.Flag("address", "The address to listen on").Default(defaultAddress).StringVar(&config.Address)
	kingpin.Flag("metrics-address", "The metrics address for Prometheus").Default(defaultMetricsAddress).StringVar(&config.MetricsAddress)
	kingpin.Flag("availability-zones", "List of AWS availability zones.").Required().StringVar(&config.AvailabilityZones)
	kingpin.Flag("tls-cert-file", "File containing the certificate for HTTPS").Required().StringVar(&config.CertFile)
	kingpin.Flag("tls-key-file", "File containing the private key for HTTPS").Required().StringVar(&config.KeyFile)

	kingpin.Parse()

	return config, nil
}
