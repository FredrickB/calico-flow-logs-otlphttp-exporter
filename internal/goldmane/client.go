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

const (
	BACKOFF = 5 * time.Second
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

func (client *GoldmaneClient) StreamFlows(context context.Context, done chan<- bool) (<-chan *pb.Flow, error) {
	data := make(chan *pb.Flow)

	go func() {
		// Begin the outer loop handling initialization
		for {
			stream, err := client.flowCollectorClient.Stream(context, &pb.FlowStreamRequest{})

			if err != nil {
				log.Printf("Failed to create stream: %s", err)
				time.Sleep(BACKOFF)
				continue
			}

			log.Printf("Stream created")

			for {
				select {
				case <-context.Done():
					log.Println("Context done, closing channel, terminating goroutine")
					close(data)
					done <- true
					return
				default:
					var flowResult pb.FlowResult
					err := stream.RecvMsg(&flowResult)

					if errors.Is(err, io.EOF) {
						// the sender has nothing more to send.
						// Terminate execution
						log.Printf("Received io.EOF, closing channel")
						close(data)
						done <- true
						return
					}

					if err != nil {
						if status, ok := status.FromError(err); ok {
							log.Printf("Failed to receive message. Status: %s. Reconnecting", status.Code())
							time.Sleep(BACKOFF)
							// break out of the select to trigger
							// reinitialization of stream
							break
						} else {
							// this is an error we don't recognize
							// and presume we can't handle.
							// Terminate execution
							log.Printf("Unknown error: %s. Terminating", err)
							close(data)
							done <- true
							return
						}
					}

					// pass data from Goldmane and proceed to
					// next iteration to block and wait for
					// more data from Goldmane
					data <- flowResult.Flow
					continue
				}
				// break out of the loop to reinitialize
				// the stream, only happens if there is
				// an error received which we believe we
				// can handle
				break
			}
		}
	}()

	return data, nil
}
