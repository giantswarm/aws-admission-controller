package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyson/certman"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/giantswarm/aws-admission-controller/v2/config"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2/awscluster"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2/awscontrolplane"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2/awsmachinedeployment"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2/cluster"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2/g8scontrolplane"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2/machinedeployment"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha2/networkpool"
	v1alpha3awscluster "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/awscluster"
	v1alpha3awscontrolplane "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/awscontrolplane"
	v1alpha3awsmachinedeployment "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/awsmachinedeployment"
	v1alpha3cluster "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/cluster"
	v1alpha3g8scontrolplane "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/g8scontrolplane"
	v1alpha3machinedeployment "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/machinedeployment"
	v1alpha3networkpool "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/networkpool"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/mutator"
	"github.com/giantswarm/aws-admission-controller/v2/pkg/validator"
)

func main() {
	config, err := config.Parse()
	if err != nil {
		panic(microerror.JSON(err))
	}

	// Setup handler for mutating webhook
	awsclusterMutator, err := awscluster.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	awscontrolplaneMutator, err := awscontrolplane.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	awsmachinedeploymentMutator, err := awsmachinedeployment.NewMutator(config)
	if err != nil {
		log.Fatalf("Unable to create G8s Control Plane admitter: %v", err)
	}

	clusterMutator, err := cluster.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	g8scontrolplaneMutator, err := g8scontrolplane.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	machinedeploymentMutator, err := machinedeployment.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3awsclusterMutator, err := v1alpha3awscluster.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3awscontrolplaneMutator, err := v1alpha3awscontrolplane.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3awsmachinedeploymentMutator, err := v1alpha3awsmachinedeployment.NewMutator(config)
	if err != nil {
		log.Fatalf("Unable to create G8s Control Plane admitter: %v", err)
	}

	v1alpha3clusterMutator, err := v1alpha3cluster.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3g8scontrolplaneMutator, err := v1alpha3g8scontrolplane.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3machinedeploymentMutator, err := v1alpha3machinedeployment.NewMutator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	// Setup handler for validating webhook
	awsclusterValidator, err := awscluster.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	awscontrolplaneValidator, err := awscontrolplane.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	awsmachinedeploymentValidator, err := awsmachinedeployment.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	clusterValidator, err := cluster.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	g8scontrolplaneValidator, err := g8scontrolplane.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	machinedeploymentValidator, err := machinedeployment.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	networkPoolValidator, err := networkpool.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3awsclusterValidator, err := v1alpha3awscluster.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3awscontrolplaneValidator, err := v1alpha3awscontrolplane.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3awsmachinedeploymentValidator, err := v1alpha3awsmachinedeployment.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3clusterValidator, err := v1alpha3cluster.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3g8scontrolplaneValidator, err := v1alpha3g8scontrolplane.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3machinedeploymentValidator, err := v1alpha3machinedeployment.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	v1alpha3networkPoolValidator, err := v1alpha3networkpool.NewValidator(config)
	if err != nil {
		panic(microerror.JSON(err))
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	// v1alpha2
	handler.Handle("/mutate/v1alpha2/awscluster", mutator.Handler(awsclusterMutator))
	handler.Handle("/mutate/v1alpha2/awsmachinedeployment", mutator.Handler(awsmachinedeploymentMutator))
	handler.Handle("/mutate/v1alpha2/awscontrolplane", mutator.Handler(awscontrolplaneMutator))
	handler.Handle("/mutate/v1alpha2/cluster", mutator.Handler(clusterMutator))
	handler.Handle("/mutate/v1alpha2/g8scontrolplane", mutator.Handler(g8scontrolplaneMutator))
	handler.Handle("/mutate/v1alpha2/machinedeployment", mutator.Handler(machinedeploymentMutator))
	handler.Handle("/validate/v1alpha2/awscluster", validator.Handler(awsclusterValidator))
	handler.Handle("/validate/v1alpha2/awscontrolplane", validator.Handler(awscontrolplaneValidator))
	handler.Handle("/validate/v1alpha2/awsmachinedeployment", validator.Handler(awsmachinedeploymentValidator))
	handler.Handle("/validate/v1alpha2/cluster", validator.Handler(clusterValidator))
	handler.Handle("/validate/v1alpha2/g8scontrolplane", validator.Handler(g8scontrolplaneValidator))
	handler.Handle("/validate/v1alpha2/machinedeployment", validator.Handler(machinedeploymentValidator))
	handler.Handle("/validate/v1alpha2/networkpool", validator.Handler(networkPoolValidator))

	// v1alpha3
	handler.Handle("/mutate/v1alpha3/awscluster", mutator.Handler(v1alpha3awsclusterMutator))
	handler.Handle("/mutate/v1alpha3/awsmachinedeployment", mutator.Handler(v1alpha3awsmachinedeploymentMutator))
	handler.Handle("/mutate/v1alpha3/awscontrolplane", mutator.Handler(v1alpha3awscontrolplaneMutator))
	handler.Handle("/mutate/v1alpha3/cluster", mutator.Handler(v1alpha3clusterMutator))
	handler.Handle("/mutate/v1alpha3/g8scontrolplane", mutator.Handler(v1alpha3g8scontrolplaneMutator))
	handler.Handle("/mutate/v1alpha3/machinedeployment", mutator.Handler(v1alpha3machinedeploymentMutator))
	handler.Handle("/validate/v1alpha3/awscluster", validator.Handler(v1alpha3awsclusterValidator))
	handler.Handle("/validate/v1alpha3/awscontrolplane", validator.Handler(v1alpha3awscontrolplaneValidator))
	handler.Handle("/validate/v1alpha3/awsmachinedeployment", validator.Handler(v1alpha3awsmachinedeploymentValidator))
	handler.Handle("/validate/v1alpha3/cluster", validator.Handler(v1alpha3clusterValidator))
	handler.Handle("/validate/v1alpha3/g8scontrolplane", validator.Handler(v1alpha3g8scontrolplaneValidator))
	handler.Handle("/validate/v1alpha3/machinedeployment", validator.Handler(v1alpha3machinedeploymentValidator))
	handler.Handle("/validate/v1alpha3/networkpool", validator.Handler(v1alpha3networkPoolValidator))

	handler.HandleFunc("/healthz", healthCheck)
	metrics := http.NewServeMux()
	metrics.Handle("/metrics", promhttp.Handler())

	go serveMetrics(config, metrics)
	serveTLS(config, handler)
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("ok"))
	if err != nil {
		panic(microerror.JSON(err))
	}
}

func serveTLS(config config.Config, handler http.Handler) {
	cm, err := certman.New(config.CertFile, config.KeyFile)
	if err != nil {
		panic(microerror.JSON(err))
	}
	if err := cm.Watch(); err != nil {
		panic(microerror.JSON(err))
	}

	server := &http.Server{
		Addr:    config.Address,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: cm.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(microerror.JSON(err))
		}
	}()

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}

func serveMetrics(config config.Config, handler http.Handler) {
	server := &http.Server{
		Addr:    config.MetricsAddress,
		Handler: handler,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(microerror.JSON(err))
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}
