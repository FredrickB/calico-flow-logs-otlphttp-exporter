package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/util"
)

const (
	CA_CERT_PATH_ENV                   string        = "CA_CERT_PATH"
	PRIVATE_KEY_PATH_ENV               string        = "PRIVATE_KEY_PATH"
	PUBLIC_CERT_PATH_ENV               string        = "PUBLIC_CERT_PATH"
	GOLDMANE_HOST_ENV                  string        = "GOLDMANE_HOST"
	RECONNECT_WAIT_TIME_IN_SECONDS_ENV string        = "RECONNECT_WAIT_TIME_IN_SECONDS"
	DEFAULT_RECONNECT_WAIT_TIME        time.Duration = 5 * time.Second
	PACKAGE_NAME                       string        = "calico-flow-logs-otlphttp-exporter"
	SERVICE_NAME                       string        = "calico-flow-logs-otlphttp-exporter"
	SERVICE_VERSION                    string        = "REPLACED_DURING_BUILD"
)

var (
	caCertFilePath, caCertSet                                   = os.LookupEnv(CA_CERT_PATH_ENV)
	publicCertPath, publicCertSet                               = os.LookupEnv(PUBLIC_CERT_PATH_ENV)
	privateKeyPath, privateKeySet                               = os.LookupEnv(PRIVATE_KEY_PATH_ENV)
	goldmaneHost, goldmaneHostSet                               = os.LookupEnv(GOLDMANE_HOST_ENV)
	reconnectSecondsStringValue, reconnectSecondsStringValueSet = os.LookupEnv(RECONNECT_WAIT_TIME_IN_SECONDS_ENV)
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
	if reconnectSecondsStringValueSet {
		util.LogEnvironmentVariable(RECONNECT_WAIT_TIME_IN_SECONDS_ENV, reconnectSecondsStringValue)
	}

	tlsConfig, err := goldmane.NewTLSConfig(caCertFilePath, publicCertPath, privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to construct TLS certificate: %s", err)
	}
	connection, err := grpc.NewClient(goldmaneHost, grpc.WithTransportCredentials(tlsConfig))
	if err != nil {
		log.Fatalf("Cannot make a connection to Goldmane on host: %s. Error: %s", goldmaneHost, err)
	}
	client := goldmane.NewClient(goldmaneHost, pb.NewFlowsClient(connection))

	context, cancel := context.WithCancel(context.Background())

	loggerProvider, err := otlp.NewLoggerProvider(context, SERVICE_NAME, SERVICE_VERSION)
	if err != nil {
		log.Fatalf("Error while creating loggerprovider: %s", err)
	}
	processor := otlp.NewProcessor(context, PACKAGE_NAME, loggerProvider)
	otlpLogger := otlp.NewLogger(processor)

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	reconnectWaitTime := util.ParseSecondsStringValue(reconnectSecondsStringValue, DEFAULT_RECONNECT_WAIT_TIME)
	log.Printf("Reconnect wait time set to: %s", reconnectWaitTime)

	streamClosed, err := util.StartLogStreaming(context, client, otlpLogger, reconnectWaitTime)
	if err != nil {
		log.Fatalf("Failed to start streaming logs: %s", err)
	}

	done := make(chan bool)
	go monitor(context, client, signals, streamClosed, done, cancel, otlpLogger, connection)
	<-done

	log.Println("Program terminated")
}

// monitor execution and cleanup when
// stream is closed or termination
// signal is received
func monitor(
	context context.Context,
	client goldmane.GoldmaneApi,
	signals chan os.Signal,
	streamClosed chan bool,
	done chan bool,
	cancel func(),
	logger otlp.OtlpLogger,
	connection *grpc.ClientConn,
) {
	for {
		select {
		case <-context.Done():
			log.Println("Context done")
			cleanup(context, client, logger, cancel, connection)
			done <- true
			return
		case <-signals:
			log.Println("Termination signal received")
			cleanup(context, client, logger, cancel, connection)
			done <- true
			return
		case <-streamClosed:
			log.Println("Stream closed")
			cleanup(context, client, logger, cancel, connection)
			done <- true
			return
		}
	}
}

func cleanup(context context.Context, client goldmane.GoldmaneApi, logger otlp.OtlpLogger, cancel func(), connection *grpc.ClientConn) {
	log.Println("Triggering cleanup...")
	util.Cleanup(context, client, logger, connection)
	cancel()
	log.Println("Cleanup finished")
}
