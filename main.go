package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/giantswarm/g8s-admission-controller/config"
	"github.com/giantswarm/g8s-admission-controller/pkg/admission"
	"github.com/giantswarm/g8s-admission-controller/pkg/g8scontrolplane"
)

func main() {
	cfg := config.Parse()

	g8scontrolplaneAdmitter, err := g8scontrolplane.NewAdmitter(&cfg.G8sControlPaneConfig)
	if err != nil {
		log.Fatalf("Unable to create Pod admitter: %v", err)
	}

	handler := http.NewServeMux()
	handler.Handle("/g8scontrolplane", admission.Handler(g8scontrolplaneAdmitter))
	handler.HandleFunc("/healthz", healthCheck)

	serve(cfg, handler)
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("ok"))
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
		server.Shutdown(context.Background())
	}()

	err := server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}
}
