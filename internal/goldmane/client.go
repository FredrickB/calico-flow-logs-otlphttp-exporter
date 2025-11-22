package goldmane

import (
	"context"
	"io"
	"log"

	pb "github.com/FredrickB/calico-flow-logs-loki-exporter/v2/proto"
	"google.golang.org/grpc"
)

type GoldmaneClient struct {
	connection          *grpc.ClientConn
	flowCollectorClient pb.FlowsClient
}

func NewClient(host, caCertificateFilepath, publicCertFilepath, privateKeyFilepath string) (*GoldmaneClient, error) {
	tlsConfig, err := newTLSConfig(caCertificateFilepath, publicCertFilepath, privateKeyFilepath)
	if err != nil {
		log.Printf("Failed to construct TLS certificate: %s", err)
		return nil, err
	}

	connection, err := grpc.NewClient(host, grpc.WithTransportCredentials(tlsConfig))
	if err != nil {
		log.Printf("Cannot make a connection to Goldmane on host: %s. Error: %s", host, err)
		return nil, err
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
		log.Printf("Failed to get flows, error: %s. Returning empty list", err)
		return &pb.FlowListResult{}, err
	}

	return list, nil
}

func (client *GoldmaneClient) StreamFlows(context context.Context, receiver chan<- *pb.Flow) {
	stream, err := client.flowCollectorClient.Stream(context, &pb.FlowStreamRequest{})
	if err != nil {
		log.Printf("Failed to create streaming client for flow api: %s", err)
		close(receiver)
	}
	for {
		var flowResult pb.FlowResult
		err := stream.RecvMsg(&flowResult)
		if err == io.EOF {
			log.Printf("Received EOF, closing channel")
			close(receiver)
		}
		if err != nil {
			log.Fatalf("Failed to receive message %s", err)
			close(receiver)
		}
		receiver <- flowResult.Flow
	}
}

func (client *GoldmaneClient) Close() error {
	return client.connection.Close()
}
