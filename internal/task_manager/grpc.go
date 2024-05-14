package task_manager

import (
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	core.Comm
	tm *taskManager
}

func (c *grpcServer) EventChannel(srv pb.Core_EventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method EventChannel not implemented")
}

func GetOperatorNodeClient(opId string) *pb.Core_EventChannelClient {
	// pending ....
	return nil
}

func NewSingleEventReq(data []byte, sourceTaskId string, targetTaskId string) *pb.EventChannelReq {
	event := &pb.Event{
		EventTime:      time.Now().Unix(),
		Data:           data,
		SourceOpTaskId: sourceTaskId,
		TargetOpTaskId: targetTaskId,
	}
	events := make([]*pb.Event, 0)
	events = append(events, event)
	return &pb.EventChannelReq{
		Events: events,
	}
}

func NewCoupleEventsReq() *pb.EventChannelReq {
	return nil
}
