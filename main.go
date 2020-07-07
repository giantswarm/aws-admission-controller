package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/giantswarm/admission-controller/config"
	"github.com/giantswarm/admission-controller/pkg/admission"
	"github.com/giantswarm/admission-controller/pkg/awsmachinedeployment"
	"github.com/giantswarm/admission-controller/pkg/g8scontrolplane"
)

func main() {
	cfg := config.Parse()

	g8scontrolplaneAdmitter, err := g8scontrolplane.NewAdmitter(&cfg.G8sControlPaneConfig)
	if err != nil {
		log.Fatalf("Unable to create G8s Control Plane admitter: %v", err)
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

	serve(cfg, handler)
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("ok"))
	if err != nil {
		log.Fatalf("Healthcheck Error: %v", err)
	}
}

func serve(cfg *config.Config, handler http.Handler) {
	server := &http.Server{
		Addr:    cfg.Address,
		Handler: handler,
	}

	log.Infof("Starting server on %s", cfg.Address)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("Shutdown Error: %v", err)
		}
	}()

	err := server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("Listen Error: %v", err)
		}
	}
}
