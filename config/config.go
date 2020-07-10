package config

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/admission-controller/pkg/g8scontrolplane"
)

const (
	defaultAddress = ":8080"
)

type Config struct {
	CertFile          string
	KeyFile           string
	Address           string
	AvailabilityZones string

	G8sControlPlane g8scontrolplane.Config
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
	kingpin.Flag("availability-zones", "List of AWS availability zones.").Required().StringVar(&result.G8sControlPlane.ValidAvailabilityZones)

	// add logger to each admission handler
	result.G8sControlPlane.Logger = newLogger

	kingpin.Parse()
	return result, nil
}
