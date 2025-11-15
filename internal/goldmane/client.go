package goldmane

import (
	"context"
	"io"
	"log"

	pb "github.com/FredrickB/calico-flow-logs-loki-exporter/v2/proto"
	"google.golang.org/grpc"
)

// TODO: Map remaining fields from pb.FlowResult.Flow.Key
type GoldmaneFlowKey struct {
	SourceName           string
	SourceNamespace      string
	DestName             string
	DestNamespace        string
	DestPort             int64
	DestServiceName      string
	DestServiceNamespace string
	DestServicePortName  string
	DestServicePort      int64
	Proto                string
}

type GoldmaneFlow struct {
	Key          GoldmaneFlowKey
	StartTime    int64
	EndTime      int64
	SourceLabels []string
	DestLabels   []string
	PacketsIn    int64
	PacketsOut   int64
}

type GoldmaneClient struct {
	connection          *grpc.ClientConn
	flowCollectorClient pb.FlowsClient
}

func NewClient(host, caCertificateFilepath, publicCertFilepath, privateKeyFilepath string) (*GoldmaneClient, error) {
	tlsConfig, err := NewTLSConfig(caCertificateFilepath, publicCertFilepath, privateKeyFilepath)
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

func (client *GoldmaneClient) StreamFlows(context context.Context, receiver chan<- GoldmaneFlow) {
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
		receiver <- mapFlowResultToGoldmaneFlow(&flowResult)
	}
}

func (client *GoldmaneClient) Close() error {
	return client.connection.Close()
}

func mapFlowResultToGoldmaneFlow(flowResult *pb.FlowResult) GoldmaneFlow {
	return GoldmaneFlow{
		Key: GoldmaneFlowKey{
			SourceName:           flowResult.Flow.Key.SourceName,
			SourceNamespace:      flowResult.Flow.Key.SourceNamespace,
			DestName:             flowResult.Flow.Key.DestName,
			DestNamespace:        flowResult.Flow.Key.DestNamespace,
			DestPort:             flowResult.Flow.Key.DestPort,
			DestServiceName:      flowResult.Flow.Key.DestServiceName,
			DestServiceNamespace: flowResult.Flow.Key.DestServiceNamespace,
			DestServicePortName:  flowResult.Flow.Key.DestServicePortName,
			DestServicePort:      flowResult.Flow.Key.DestServicePort,
			Proto:                flowResult.Flow.Key.Proto,
		},
		StartTime:    flowResult.Flow.StartTime,
		EndTime:      flowResult.Flow.EndTime,
		SourceLabels: flowResult.Flow.SourceLabels,
		DestLabels:   flowResult.Flow.DestLabels,
		PacketsIn:    flowResult.Flow.PacketsIn,
		PacketsOut:   flowResult.Flow.PacketsOut,
	}
}
