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
	awscluster "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/awscluster"
	awscontrolplane "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/awscontrolplane"
	awsmachinedeployment "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/awsmachinedeployment"
	cluster "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/cluster"
	g8scontrolplane "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/g8scontrolplane"
	machinedeployment "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/machinedeployment"
	networkpool "github.com/giantswarm/aws-admission-controller/v2/pkg/aws/v1alpha3/networkpool"
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

	// Here we register our endpoints.
	handler := http.NewServeMux()

	//
	handler.Handle("/mutate/v1alpha3/awscluster", mutator.Handler(awsclusterMutator))
	handler.Handle("/mutate/v1alpha3/awsmachinedeployment", mutator.Handler(awsmachinedeploymentMutator))
	handler.Handle("/mutate/v1alpha3/awscontrolplane", mutator.Handler(awscontrolplaneMutator))
	handler.Handle("/mutate/v1alpha3/cluster", mutator.Handler(clusterMutator))
	handler.Handle("/mutate/v1alpha3/g8scontrolplane", mutator.Handler(g8scontrolplaneMutator))
	handler.Handle("/mutate/v1alpha3/machinedeployment", mutator.Handler(machinedeploymentMutator))
	handler.Handle("/validate/v1alpha3/awscluster", validator.Handler(awsclusterValidator))
	handler.Handle("/validate/v1alpha3/awscontrolplane", validator.Handler(awscontrolplaneValidator))
	handler.Handle("/validate/v1alpha3/awsmachinedeployment", validator.Handler(awsmachinedeploymentValidator))
	handler.Handle("/validate/v1alpha3/cluster", validator.Handler(clusterValidator))
	handler.Handle("/validate/v1alpha3/g8scontrolplane", validator.Handler(g8scontrolplaneValidator))
	handler.Handle("/validate/v1alpha3/machinedeployment", validator.Handler(machinedeploymentValidator))
	handler.Handle("/validate/v1alpha3/networkpool", validator.Handler(networkPoolValidator))

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
