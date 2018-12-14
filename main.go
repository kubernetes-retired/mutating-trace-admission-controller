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
	"go.opencensus.io/trace"
)

func main() {

	// read configuration location from command line arg
	var configPath string
	flag.StringVar(&configPath, "configPath", DefaultConfigPath, "Path that points to the YAML configuration for this webhook.")
	flag.Parse()

	// parse and validate configuration
	config := Config{}

	ok, err := ParseConfigFromPath(&config, configPath)
	if !ok {
		glog.Errorf("configuration parse failed with error: %v", err)
		return
	}

	ok, err = config.Validate()
	if !ok {
		glog.Errorf("configuration validation failed with error: %v", err)
		return
	}

	// configure global tracer
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(config.Trace.SampleRate)})

	// configure certificates
	pair, err := tls.LoadX509KeyPair("/etc/webhook/certs/cert.pem", "/etc/webhook/certs/key.pem")
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	whsvr := &WebhookServer{
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", 443),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.serve)
	whsvr.server.Handler = mux

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
