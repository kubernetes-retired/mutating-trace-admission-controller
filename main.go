package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

func main() {
	var parameters WebhookServerParameters
	var traceConfig TraceConfig

	// get command line parameters
	flag.Float64Var(&traceConfig.samplingRate, "rate", 1.00, "Float between 0.0 and 1.0 inclusive which sets the probability that any given span context is sampled")
	flag.IntVar(&parameters.port, "port", 443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	whsvr := &WebhookServer{
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.serve)
	whsvr.server.Handler = mux

	// pass configuration through here
	err = ConfigureTracing(&traceConfig)
	if err != nil {
		glog.Errorf("could not validate passed tracing configuration: %v", err)
		return
	}

	// begin webhook server
	go func() {
		if err := whsvr.server.ListenAndServeTLS("", ""); err != nil {
			glog.Fatalf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Errorf("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsvr.server.Shutdown(context.Background())
}
