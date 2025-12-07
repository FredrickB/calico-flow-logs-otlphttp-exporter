package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/util"
)

const (
	CA_CERT_PATH_ENV     string = "CA_CERT_PATH"
	PRIVATE_KEY_PATH_ENV string = "PRIVATE_KEY_PATH"
	PUBLIC_CERT_PATH_ENV string = "PUBLIC_CERT_PATH"
	GOLDMANE_HOST_ENV    string = "GOLDMANE_HOST"
	PACKAGE_NAME         string = "calico-flow-logs-otlphttp-exporter"
	SERVICE_NAME         string = "calico-flow-logs-otlphttp-exporter"
	SERVICE_VERSION      string = "REPLACED_DURING_BUILD"
)

var (
	caCertFilePath, caCertSet     = os.LookupEnv(CA_CERT_PATH_ENV)
	publicCertPath, publicCertSet = os.LookupEnv(PUBLIC_CERT_PATH_ENV)
	privateKeyPath, privateKeySet = os.LookupEnv(PRIVATE_KEY_PATH_ENV)
	goldmaneHost, goldmaneHostSet = os.LookupEnv(GOLDMANE_HOST_ENV)
)

func main() {
	appLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(appLogger)

	if !caCertSet || !privateKeySet || !publicCertSet || !goldmaneHostSet {
		log.Fatalf("One of the following environment variables is not set: %s, %s, %s, %s. All of these need to be set",
			CA_CERT_PATH_ENV, PRIVATE_KEY_PATH_ENV, PUBLIC_CERT_PATH_ENV, GOLDMANE_HOST_ENV)
	}

	util.LogEnvironmentVariable(CA_CERT_PATH_ENV, caCertFilePath)
	util.LogEnvironmentVariable(PRIVATE_KEY_PATH_ENV, privateKeyPath)
	util.LogEnvironmentVariable(PUBLIC_CERT_PATH_ENV, publicCertPath)
	util.LogEnvironmentVariable(GOLDMANE_HOST_ENV, goldmaneHost)

	client, err := goldmane.NewClient(
		goldmaneHost,
		caCertFilePath,
		publicCertPath,
		privateKeyPath,
	)
	if err != nil {
		log.Fatalf("Error while creating Goldmane client: %s", err)
	}

	context, cancel := context.WithCancel(context.Background())

	otlpLogger, err := otlp.NewLogger(context, PACKAGE_NAME, SERVICE_NAME, SERVICE_VERSION)
	if err != nil {
		log.Fatalf("Error while creating logger: %s", err)
	}

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	log.Println("Start streaming logs from Goldmane...")
	streamClosed, err := util.StartLogStreaming(context, client, otlpLogger)
	if err != nil {
		log.Fatalf("Failed to start streaming logs: %s", err)
	}

	done := make(chan bool)
	go monitor(context, client, signals, streamClosed, done, cancel, otlpLogger)
	<-done

	log.Println("Program terminated")
}

// monitor execution and cleanup when
// stream is closed or termination
// signal is received
func monitor(
	context context.Context,
	client *goldmane.GoldmaneClient,
	signals chan os.Signal,
	streamClosed chan bool,
	done chan bool,
	cancel func(),
	logger *otlp.Logger,
) {
	for {
		select {
		case <-context.Done():
			log.Println("Context done")
			cleanup(context, client, logger, cancel)
			done <- true
			return
		case <-signals:
			log.Println("Termination signal received")
			cleanup(context, client, logger, cancel)
			done <- true
			return
		case <-streamClosed:
			log.Println("Stream closed")
			cleanup(context, client, logger, cancel)
			done <- true
			return
		}
	}
}

func cleanup(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger, cancel func()) {
	log.Println("Triggering cleanup...")
	util.Cleanup(context, client, logger)
	cancel()
	log.Println("Cleanup finished")
}
