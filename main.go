package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/admission-controller/config"
	"github.com/giantswarm/admission-controller/pkg/admission"
	"github.com/giantswarm/admission-controller/pkg/g8scontrolplane"
)

func main() {
	config := config.Parse()

	g8scontrolplaneAdmitter, err := g8scontrolplane.NewAdmitter(config.G8sControlPlane)
	if err != nil {
		panic(fmt.Errorf("Unable to create G8s Control Plane admitter: %v", err))
	}

	handler := http.NewServeMux()
	handler.Handle("/g8scontrolplane", admission.Handler(g8scontrolplaneAdmitter))
	handler.HandleFunc("/healthz", healthCheck)

	serve(config, handler)
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("ok"))
	if err != nil {
		panic(err)
	}
}

func serve(config config.Config, handler http.Handler) {
	server := &http.Server{
		Addr:    config.Address,
		Handler: handler,
	}

	config.Logger.Log("level", "debug", "message", fmt.Sprintf("starting server on %s", config.Address))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	err := server.ListenAndServeTLS(config.CertFile, config.KeyFile)
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
