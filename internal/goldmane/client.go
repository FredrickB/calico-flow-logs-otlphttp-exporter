package goldmane

import (
	pb "github.com/FredrickB/calico-flow-logs-loki-exporter/v2/proto"
	"google.golang.org/grpc"
)

func NewClient() {
	pb.NewFlowCollectorClient(&grpc.ClientConn{})
}
