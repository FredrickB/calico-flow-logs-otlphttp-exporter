package goldmane

import (
	"context"
	"testing"
	"time"

	"github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/mocks"
	pb "github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/gen/protos"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
)

var (
	flowResult = &pb.FlowResult{
		Flow: &pb.Flow{
			SourceLabels: []string{"test=true"},
		},
	}
)

type MockGrpcServerStreamingClient[Res pb.FlowResult] struct {
	MockClientStream
}

func (MockGrpcServerStreamingClient[Res]) Recv() (*pb.FlowResult, error) { return flowResult, nil }

type MockClientStream struct{}

func (MockClientStream) Header() (metadata.MD, error)  { return nil, nil }
func (MockClientStream) Trailer() metadata.MD          { return nil }
func (MockClientStream) CloseSend() error              { return nil }
func (MockClientStream) Context() context.Context      { return nil }
func (MockClientStream) SendMsg(m any) error           { return nil }
func (MockClientStream) RecvMsg(m any) error           { return nil }
func (MockClientStream) Recv() (*pb.FlowResult, error) { return flowResult, nil }

func TestStreamFlowsPassesData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlowsClient := mocks.NewMockFlowsClient(ctrl)

	client := NewClient("goldmane:7443", mockFlowsClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockDoneCh := make(chan bool, 2)
	reconnectWaitTime := 2 * time.Second

	mockGrpcServerStreamingClient := MockGrpcServerStreamingClient[pb.FlowResult]{}
	mockFlowsClient.EXPECT().Stream(ctx, &pb.FlowStreamRequest{}).Return(mockGrpcServerStreamingClient, nil)

	dataCh, err := client.StreamFlows(ctx, mockDoneCh, reconnectWaitTime)
	if err != nil {
		t.Error("err should be nil")
	}
	if dataCh == nil {
		t.Error("data channel should not be nil")
	}

	receivedFlow := <-dataCh
	if receivedFlow != flowResult.Flow {
		t.Errorf("expected %s, got %s", flowResult.Flow, receivedFlow)
	}
}
