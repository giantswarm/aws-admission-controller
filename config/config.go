package config

import (
	"github.com/giantswarm/admission-controller/pkg/g8scontrolplane"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultAddress = ":8080"
)

type Config struct {
	CertFile          string
	KeyFile           string
	Address           string
	AvailabilityZones string

	G8sControlPaneConfig g8scontrolplane.AdmitterConfig
}

func Parse() *Config {
	result := &Config{}
	kingpin.Flag("tls-cert-file", "File containing the certificate for HTTPS").Required().StringVar(&result.CertFile)
	kingpin.Flag("tls-key-file", "File containing the private key for HTTPS").Required().StringVar(&result.KeyFile)
	kingpin.Flag("address", "The address to listen on").Default(defaultAddress).StringVar(&result.Address)
	kingpin.Flag("availability-zones", "List of AWS availability zones.").Required().StringVar(&result.G8sControlPaneConfig.ValidAvailabilityZones)

	kingpin.Parse()
	return result
}
