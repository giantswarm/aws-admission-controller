package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/admission-controller/config"
	"github.com/giantswarm/admission-controller/pkg/admission"
	"github.com/giantswarm/admission-controller/pkg/awsmachinedeployment"
	"github.com/giantswarm/admission-controller/pkg/g8scontrolplane"
)

func main() {
	config, err := config.Parse()
	if err != nil {
		panic(microerror.JSON(err))
	}

	g8scontrolplaneAdmitter, err := g8scontrolplane.NewAdmitter(config.G8sControlPlane)
	if err != nil {
		panic(microerror.JSON(err))
	}

	awsMachineDeploymentAdmitter, err := awsmachinedeployment.NewAdmitter(nil)
	if err != nil {
		log.Fatalf("Unable to create G8s Control Plane admitter: %v", err)
	}

	// Here we register our endpoints.
	handler := http.NewServeMux()
	handler.Handle("/awsmachinedeployment", admission.Handler(awsMachineDeploymentAdmitter))
	handler.Handle("/g8scontrolplane", admission.Handler(g8scontrolplaneAdmitter))
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
