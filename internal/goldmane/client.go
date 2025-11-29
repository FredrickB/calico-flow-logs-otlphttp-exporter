package goldmane

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"google.golang.org/grpc"
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

func (client *GoldmaneClient) GetFlow(context context.Context) (*pb.FlowListResult, error) {
	list, err := client.flowCollectorClient.List(
		context,
		&pb.FlowListRequest{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get flows, error: %s. Returning empty list", err)
	}

	return list, nil
}

func (client *GoldmaneClient) StreamFlows(context context.Context) (<-chan *pb.Flow, error) {
	stream, err := client.flowCollectorClient.Stream(context, &pb.FlowStreamRequest{})

	if err != nil {
		return nil, fmt.Errorf("failed to create streaming client for flow api %s", err)
	}

	data := make(chan *pb.Flow)

	go func() {
		for {
			select {
			case <-context.Done():
				log.Println("Context done, closing channel, terminating goroutine")
				close(data)
				return
			default:
				var flowResult pb.FlowResult
				err := stream.RecvMsg(&flowResult)
				if err == io.EOF {
					log.Printf("Received EOF, closing channel")
					close(data)
				}
				if err != nil {
					log.Printf("Failed to receive message. Error: %s, closing channel", err)
					close(data)
				}
				data <- flowResult.Flow
			}
		}
	}()

	return data, nil
}

func (client *GoldmaneClient) Close() error {
	return client.connection.Close()
}
