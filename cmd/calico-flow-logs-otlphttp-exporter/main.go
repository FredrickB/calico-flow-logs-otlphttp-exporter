package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/proto"
)

const (
	CA_CERT_PATH_ENV     string = "CA_CERT_PATH"
	PRIVATE_KEY_PATH_ENV string = "PRIVATE_KEY_PATH"
	PUBLIC_CERT_PATH_ENV string = "PUBLIC_CERT_PATH"
	GOLDMANE_HOST_ENV    string = "GOLDMANE_HOST"
	PACKAGE_NAME         string = "calico-flow-logs-otlphttp-exporter"
	SERVICE_NAME         string = "calico-flow-logs-otlphttp-exporter"
	SERVICE_VERSION      string = "0.0.1"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get path to certs required for Goldmane communication
	caCertFilePath, caCertSet := os.LookupEnv(CA_CERT_PATH_ENV)
	publicCertPath, publicCertSet := os.LookupEnv(PUBLIC_CERT_PATH_ENV)
	privateKeyPath, privateKeySet := os.LookupEnv(PRIVATE_KEY_PATH_ENV)
	goldmaneHost, goldmaneHostSet := os.LookupEnv(GOLDMANE_HOST_ENV)

	if !caCertSet || !privateKeySet || !publicCertSet || !goldmaneHostSet {
		log.Fatalf("One of the following environment variables is not set: %s, %s, %s, %s. All of these need to be set",
			CA_CERT_PATH_ENV, PRIVATE_KEY_PATH_ENV, PUBLIC_CERT_PATH_ENV, GOLDMANE_HOST_ENV)
	}
	log.Printf("%s set to %s", CA_CERT_PATH_ENV, caCertFilePath)
	log.Printf("%s set to %s", PRIVATE_KEY_PATH_ENV, privateKeyPath)
	log.Printf("%s set to %s", PUBLIC_CERT_PATH_ENV, publicCertPath)
	log.Printf("%s set to %s", GOLDMANE_HOST_ENV, goldmaneHost)

	// create goldmane client
	client, err := goldmane.NewClient(
		goldmaneHost,
		caCertFilePath,
		publicCertPath,
		privateKeyPath,
	)
	if err != nil {
		log.Fatalf("Error while creating Goldmane client: %s", err)
	}

	context := context.Background()
	// create logger
	logger, err := otlp.NewLogger(context, PACKAGE_NAME, SERVICE_NAME, SERVICE_VERSION)
	if err != nil {
		log.Fatalf("Error while creating logger")
	}

	// register signals to terminate application
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	go func() {
		<-signals

		// trigger cleanup
		log.Println("Termination signal received, triggering cleanup...")
		cleanup(context, client, logger)
		log.Println("Cleanup finished")

		// trigger termination
		done <- true
	}()

	log.Println("Start streaming logs from Goldmane to OTLP Log HTTP Exporter...")
	startLogStreaming(context, client, logger)

	<-done
	log.Println("Program terminated")
}

func startLogStreaming(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger) {
	data := make(chan *pb.Flow)

	go client.StreamFlows(context, data)
	go logger.ReceiveFlows(context, data)
}

func cleanup(context context.Context, client *goldmane.GoldmaneClient, logger *otlp.Logger) {
	if err := client.Close(); err != nil {
		log.Printf("Error while closing client: %s", err)
	}
	if err := logger.Close(context); err != nil {
		log.Printf("Error while shutting down loggerProvider: %s", err)
	}
}
