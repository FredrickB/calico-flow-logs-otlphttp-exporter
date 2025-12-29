package goldmane

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/mocks"
	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"go.uber.org/mock/gomock"
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

func TestStreamFlowsPassesData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowsClient := mocks.NewMockFlowsClient(ctrl)

	client := NewClient("goldmane:7443", mockFlowsClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reconnectWaitTime := 2 * time.Second

	mockGrpcServerStreamingClient := MockGrpcServerStreamingClient[pb.FlowResult]{}
	mockFlowsClient.EXPECT().Stream(ctx, &pb.FlowStreamRequest{}).
		Return(mockGrpcServerStreamingClient, nil)

	dataCh, _ := client.StreamFlows(ctx, reconnectWaitTime)

	receivedFlow := <-dataCh
	if receivedFlow != flowResult.Flow {
		t.Errorf("expected %s, got %s", flowResult.Flow, receivedFlow)
	}
}

func TestEOFErrorStopsStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowsClient := mocks.NewMockFlowsClient(ctrl)

	client := NewClient("goldmane:7443", mockFlowsClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flowResult = nil
	mockErr = io.EOF

	mockGrpcServerStreamingClient := MockGrpcServerStreamingClient[pb.FlowResult]{}
	mockFlowsClient.EXPECT().Stream(ctx, &pb.FlowStreamRequest{}).
		Return(mockGrpcServerStreamingClient, nil)

	dataCh, _ := client.StreamFlows(ctx, 2*time.Second)

	flow := <-dataCh
	if flow != nil {
		t.Errorf("flow should be nil, was: %s", flow)
	}
}

func TestUnknownGRPCErrorStopsStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowsClient := mocks.NewMockFlowsClient(ctrl)

	client := NewClient("goldmane:7443", mockFlowsClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flowResult = nil
	mockErr = ErrorNotMappedToGRPCStatus

	mockGrpcServerStreamingClient := MockGrpcServerStreamingClient[pb.FlowResult]{}
	mockFlowsClient.EXPECT().Stream(ctx, &pb.FlowStreamRequest{}).
		Return(mockGrpcServerStreamingClient, nil)

	dataCh, _ := client.StreamFlows(ctx, 2*time.Second)

	flow := <-dataCh
	if flow != nil {
		t.Errorf("flow should be nil, was: %s", flow)
	}
}

func TestKnownGRPCErrorTriggersReconnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowsClient := mocks.NewMockFlowsClient(ctrl)

	client := NewClient("goldmane:7443", mockFlowsClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flowResult = nil
	mockErr = status.Error(codes.NotFound, "some description")

	reconnectWaitTime := 2 * time.Second

	mockGrpcServerStreamingClient := MockGrpcServerStreamingClient[pb.FlowResult]{}
	mockFlowsClient.EXPECT().Stream(ctx, &pb.FlowStreamRequest{}).
		Return(mockGrpcServerStreamingClient, nil).
		AnyTimes()

	_, reconnectCh := client.StreamFlows(ctx, reconnectWaitTime)

	reconnectAttempt := <-reconnectCh
	if !errors.Is(reconnectAttempt, mockErr) {
		t.Errorf("reconnects should have been made")
	}
}
