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
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/core"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/goldmane"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/otlp"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/util"
	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/version"
)

const (
	CA_CERT_PATH_ENV                        string        = "CA_CERT_PATH"
	PRIVATE_KEY_PATH_ENV                    string        = "PRIVATE_KEY_PATH"
	PUBLIC_CERT_PATH_ENV                    string        = "PUBLIC_CERT_PATH"
	GOLDMANE_HOST_ENV                       string        = "GOLDMANE_HOST"
	RECONNECT_WAIT_TIME_IN_MILLISECONDS_ENV string        = "RECONNECT_WAIT_TIME_IN_MILLISECONDS"
	DEFAULT_RECONNECT_WAIT_TIME             time.Duration = 5 * time.Second
	PACKAGE_NAME                            string        = "calico-flow-logs-otlphttp-exporter"
	SERVICE_NAME                            string        = "calico-flow-logs-otlphttp-exporter"
)

var (
	caCertFilePath, caCertSet                                             = os.LookupEnv(CA_CERT_PATH_ENV)
	publicCertPath, publicCertSet                                         = os.LookupEnv(PUBLIC_CERT_PATH_ENV)
	privateKeyPath, privateKeySet                                         = os.LookupEnv(PRIVATE_KEY_PATH_ENV)
	goldmaneHost, goldmaneHostSet                                         = os.LookupEnv(GOLDMANE_HOST_ENV)
	reconnectMilliSecondsStringValue, reconnectMilliSecondsStringValueSet = os.LookupEnv(RECONNECT_WAIT_TIME_IN_MILLISECONDS_ENV)
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
	if reconnectMilliSecondsStringValueSet {
		util.LogEnvironmentVariable(RECONNECT_WAIT_TIME_IN_MILLISECONDS_ENV, reconnectMilliSecondsStringValue)
	}
	log.Printf("Version: %s", version.Version())
	log.Printf("Goldmane Protobuf Version: %s", version.GoldmaneProtobufVersion())

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

	loggerProvider, err := otlp.NewLoggerProvider(context, SERVICE_NAME, version.Version())
	if err != nil {
		log.Fatalf("Error while creating loggerprovider: %s", err)
	}
	otlpLogger := otlp.NewLogger(otlp.NewProcessor(context, PACKAGE_NAME, loggerProvider))

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	reconnectWaitTime := util.ParseMilliSecondsStringValue(reconnectMilliSecondsStringValue, DEFAULT_RECONNECT_WAIT_TIME)
	log.Printf("Reconnect wait time set to: %s", reconnectWaitTime)

	reconnects := core.Run(context, client, otlpLogger, reconnectWaitTime)

	go monitor(context, signals, cancel, otlpLogger, connection, reconnects)

	<-context.Done()

	log.Println("Program terminated")
}

// monitor execution and cleanup when
// stream is closed or termination
// signal is received
func monitor(
	context context.Context,
	signals chan os.Signal,
	cancelFunc func(),
	logger *otlp.Logger,
	connection *grpc.ClientConn,
	reconnectErrors <-chan error,
) {
	for {
		select {
		case <-context.Done():
			log.Println("Context done")
			cleanup(context, logger, cancelFunc, connection)
			return
		case <-signals:
			log.Println("Termination signal received")
			cleanup(context, logger, cancelFunc, connection)
			return
		case err := <-reconnectErrors:
			log.Printf("Reconnect attempted, error: %s", err)
		}
	}
}

func cleanup(
	context context.Context,
	logger *otlp.Logger,
	cancelFunc func(),
	connection *grpc.ClientConn,
) {
	log.Println("Triggering cleanup...")
	core.Cleanup(context, logger, connection)
	cancelFunc()
	log.Println("Cleanup finished")
}
