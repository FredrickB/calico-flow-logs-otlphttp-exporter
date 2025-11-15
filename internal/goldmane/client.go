package goldmane

import (
	"context"
	"log"

	pb "github.com/FredrickB/calico-flow-logs-loki-exporter/v2/proto"
	"google.golang.org/grpc"
)

type GoldmaneClient struct {
	connection          *grpc.ClientConn
	flowCollectorClient pb.FlowsClient
}

func NewClient(host, caCertificateFilepath, publicCertFilepath, privateKeyFilepath string) (*GoldmaneClient, error) {
	tlsConfig, err := NewTLSConfig(caCertificateFilepath, publicCertFilepath, privateKeyFilepath)
	if err != nil {
		log.Printf("Failed to construct TLS certificate: %w", err)
		return nil, err
	}

	connection, err := grpc.NewClient(host, grpc.WithTransportCredentials(tlsConfig))
	if err != nil {
		log.Printf("Cannot make a connection to Goldmane on host: %s. Error: %w", host, err)
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

func (client *GoldmaneClient) Close() error {
	return client.connection.Close()
}
