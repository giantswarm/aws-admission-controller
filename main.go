package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-admission-controller/config"
	"github.com/giantswarm/aws-admission-controller/pkg/aws/awscontrolplane"
	"github.com/giantswarm/aws-admission-controller/pkg/aws/awsmachinedeployment"
	"github.com/giantswarm/aws-admission-controller/pkg/aws/g8scontrolplane"
	"github.com/giantswarm/aws-admission-controller/pkg/mutator"
)

func main() {
	config, err := config.Parse()
	if err != nil {
		panic(microerror.JSON(err))
	}

	// Setup handler for mutating webhook
	awsMachineDeploymentMutator, err := awsmachinedeployment.NewMutator(config.AWSMachineDeployment)
	if err != nil {
		log.Fatalf("Unable to create G8s Control Plane admitter: %v", err)
	}

	awscontrolplaneMutator, err := awscontrolplane.NewMutator(config.AWSControlPlane)
	if err != nil {
		panic(microerror.JSON(err))
	}

	g8scontrolplaneMutator, err := g8scontrolplane.NewMutator(config.G8sControlPlane)
	if err != nil {
		panic(microerror.JSON(err))
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	handler.Handle("/mutate/awsmachinedeployment", mutator.Handler(awsMachineDeploymentMutator))
	handler.Handle("/mutate/awscontrolplane", mutator.Handler(awscontrolplaneMutator))
	handler.Handle("/mutate/g8scontrolplane", mutator.Handler(g8scontrolplaneMutator))
	handler.HandleFunc("/healthz", healthCheck)

	serve(config, handler)
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("ok"))
	if err != nil {
		panic(microerror.JSON(err))
	}
}

func serve(config config.Config, handler http.Handler) {
	server := &http.Server{
		Addr:    config.Address,
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

	err := server.ListenAndServeTLS(config.CertFile, config.KeyFile)
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}
