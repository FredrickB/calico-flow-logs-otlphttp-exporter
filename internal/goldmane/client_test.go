package goldmane

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	flowResult = &pb.FlowResult{
		Flow: &pb.Flow{
			SourceLabels: []string{"test=true"},
		},
	}
	mockErr                    error
	ErrorNotMappedToGRPCStatus = errors.New("error without GRPC mapping implementation")
)

type MockGrpcServerStreamingClient[Res pb.FlowResult] struct {
	MockClientStream
}

func (MockGrpcServerStreamingClient[Res]) Recv() (*pb.FlowResult, error) {
	return flowResult, mockErr
}

type MockClientStream struct{}

func (MockClientStream) Header() (metadata.MD, error) { return nil, nil }
func (MockClientStream) Trailer() metadata.MD         { return nil }
func (MockClientStream) CloseSend() error             { return nil }
func (MockClientStream) Context() context.Context     { return nil }
func (MockClientStream) SendMsg(m any) error          { return nil }
func (MockClientStream) RecvMsg(m any) error          { return nil }

type FakeFlowsClient struct {
	StreamingClient MockGrpcServerStreamingClient[pb.FlowResult]
}

func (client *FakeFlowsClient) List(ctx context.Context, in *pb.FlowListRequest, opts ...grpc.CallOption) (*pb.FlowListResult, error) {
	return nil, nil
}

func (client *FakeFlowsClient) Stream(ctx context.Context, in *pb.FlowStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pb.FlowResult], error) {
	return client.StreamingClient, nil
}

func (client *FakeFlowsClient) FilterHints(ctx context.Context, in *pb.FilterHintsRequest, opts ...grpc.CallOption) (*pb.FilterHintsResult, error) {
	return nil, nil
}

func TestStreamFlowsPassesData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockFlowsClient := FakeFlowsClient{
		StreamingClient: MockGrpcServerStreamingClient[pb.FlowResult]{},
	}

	client := NewClient("goldmane:7443", &mockFlowsClient)

	reconnectWaitTime := 2 * time.Second

	dataCh, _ := client.StreamFlows(ctx, reconnectWaitTime)

	receivedFlow := <-dataCh
	if receivedFlow != flowResult.Flow {
		t.Errorf("expected %s, got %s", flowResult.Flow, receivedFlow)
	}
}

func TestEOFErrorStopsStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockFlowsClient := FakeFlowsClient{
		StreamingClient: MockGrpcServerStreamingClient[pb.FlowResult]{},
	}

	client := NewClient("goldmane:7443", &mockFlowsClient)

	flowResult = nil
	mockErr = io.EOF

	dataCh, _ := client.StreamFlows(ctx, 2*time.Second)

	flow := <-dataCh
	if flow != nil {
		t.Errorf("flow should be nil, was: %s", flow)
	}
}

func TestUnknownGRPCErrorStopsStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockFlowsClient := FakeFlowsClient{
		StreamingClient: MockGrpcServerStreamingClient[pb.FlowResult]{},
	}

	client := NewClient("goldmane:7443", &mockFlowsClient)

	flowResult = nil
	mockErr = ErrorNotMappedToGRPCStatus

	dataCh, _ := client.StreamFlows(ctx, 2*time.Second)

	flow := <-dataCh
	if flow != nil {
		t.Errorf("flow should be nil, was: %s", flow)
	}
}

func TestKnownGRPCErrorTriggersReconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockFlowsClient := FakeFlowsClient{
		StreamingClient: MockGrpcServerStreamingClient[pb.FlowResult]{},
	}

	client := NewClient("goldmane:7443", &mockFlowsClient)

	flowResult = nil
	mockErr = status.Error(codes.NotFound, "some description")

	reconnectWaitTime := 2 * time.Second

	_, reconnectCh := client.StreamFlows(ctx, reconnectWaitTime)

	reconnectAttempt := <-reconnectCh
	if !errors.Is(reconnectAttempt, mockErr) {
		t.Errorf("reconnects should have been made")
	}
}
