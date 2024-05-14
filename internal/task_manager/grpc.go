package task_manager

import (
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"io"
	"log"
	"time"
)

type grpcServer struct {
	core.Comm
	tm *taskManager
}

func NewGrpcServer() *grpcServer {
	return &grpcServer{}
}

func (c *grpcServer) EventChannel(stream pb.Core_EventChannelServer) error {

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Read done \n")
			// read done.
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
			return err
		}
		log.Printf("Got message %v \n", in.Events)

		// todo distribute events
	}
	return nil
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
