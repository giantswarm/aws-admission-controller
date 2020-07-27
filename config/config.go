package config

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/giantswarm/admission-controller/pkg/aws/awscontrolplane"
	"github.com/giantswarm/admission-controller/pkg/aws/awsmachinedeployment"
	"github.com/giantswarm/admission-controller/pkg/aws/g8scontrolplane"
	"github.com/giantswarm/admission-controller/pkg/azureupdate"
)

const (
	defaultAddress = ":8080"
)

type Config struct {
	CertFile          string
	KeyFile           string
	Address           string
	AvailabilityZones string

	G8sControlPlane      g8scontrolplane.Config
	AWSControlPlane      awscontrolplane.Config
	AWSMachineDeployment awsmachinedeployment.Config
	AzureCluster         azureupdate.AzureClusterConfigValidatorConfig
	AzureConfig          azureupdate.AzureConfigValidatorConfig
}

func Parse() (Config, error) {
	var err error
	var result Config

	// Create a new logger that is used by all admitters.
	var newLogger micrologger.Logger
	{
		newLogger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	kingpin.Flag("tls-cert-file", "File containing the certificate for HTTPS").Required().StringVar(&result.CertFile)
	kingpin.Flag("tls-key-file", "File containing the private key for HTTPS").Required().StringVar(&result.KeyFile)
	kingpin.Flag("address", "The address to listen on").Default(defaultAddress).StringVar(&result.Address)
	kingpin.Flag("availability-zones", "List of AWS availability zones.").Required().StringVar(&result.AvailabilityZones)

	// add logger to each admission handler
	result.G8sControlPlane.Logger = newLogger
	result.AWSControlPlane.Logger = newLogger
	result.AWSMachineDeployment.Logger = newLogger
	result.AzureCluster.Logger = newLogger
	result.AzureConfig.Logger = newLogger

	kingpin.Parse()

	// add availability zones to admitter configs
	result.AWSControlPlane.ValidAvailabilityZones = result.AvailabilityZones
	result.G8sControlPlane.ValidAvailabilityZones = result.AvailabilityZones

	return result, nil
}
