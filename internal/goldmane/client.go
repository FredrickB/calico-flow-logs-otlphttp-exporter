package goldmane

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GoldmaneClient struct {
	connection          *grpc.ClientConn
	flowCollectorClient pb.FlowsClient
}

func NewClient(host, caCertificateFilepath, publicCertFilepath, privateKeyFilepath string) (*GoldmaneClient, error) {
	tlsConfig, err := newTLSConfig(caCertificateFilepath, publicCertFilepath, privateKeyFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to construct TLS certificate: %s", err)
	}

	connection, err := grpc.NewClient(host, grpc.WithTransportCredentials(tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("cannot make a connection to Goldmane on host: %s. Error: %s", host, err)
	}

	flowClientConnection := pb.NewFlowsClient(connection)

	return &GoldmaneClient{
		connection:          connection,
		flowCollectorClient: flowClientConnection,
	}, nil
}

func (client *GoldmaneClient) Close() error {
	return client.connection.Close()
}

func (client *GoldmaneClient) StreamFlows(context context.Context, done chan<- bool, reconnectWaitTime time.Duration) (<-chan *pb.Flow, error) {
	data := make(chan *pb.Flow)

	go func() {
		// Begin the outer loop handling initialization
		for {
			stream, err := client.flowCollectorClient.Stream(context, &pb.FlowStreamRequest{})

			if err != nil {
				log.Printf("Failed to create stream: %s. Sleeping for %s", err, reconnectWaitTime)
				time.Sleep(reconnectWaitTime)
				continue
			}

			log.Printf("Stream created")

			reconnect, _ := streamFlowsUntilError(context, stream, data)
			if !reconnect {
				close(data)
				done <- true
				return
			}
		}
	}()

	return data, nil
}

// start streaming flow logs to channel `data` until error is
// received from grpc. If there is a possibillity for recovery
// by reconnecting, flag it to the caller using `reconnect`
func streamFlowsUntilError(context context.Context, stream grpc.ServerStreamingClient[pb.FlowResult], data chan<- *pb.Flow) (reconnect bool, err error) {
	// start out with the assumption that
	// reconnecting won't be done for the
	// majority of errors
	reconnect = false

	log.Printf("Streaming logs...")

	for {
		select {
		case <-context.Done():
			log.Println("Context done, terminating")
			return
		default:
			var flowResult pb.FlowResult
			err = stream.RecvMsg(&flowResult)

			if errors.Is(err, io.EOF) {
				// the sender has nothing more to send, terminate
				log.Printf("Received EOF, terminating")
				return
			}

			if err != nil {
				if status, ok := status.FromError(err); ok {
					// known error, reconnect
					log.Printf("Status received: %s, trigger reconnect", status.Code())
					reconnect = true
				} else {
					// unrecognizeable error, terminate
					log.Printf("Unknown error: %s, terminating", err)
				}
				return
			}

			data <- flowResult.Flow
		}
	}
}
