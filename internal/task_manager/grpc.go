package task_manager

import (
	"bytes"
	"encoding/binary"
	"errors"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"io"
	"log/slog"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	// server push task output data
	go func() {
		defer wg.Done()
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				slog.Info("Read all client's msg done \n")
				// read done.
				break
			}
			if err != nil {
				slog.Error("Failed to receive a event : %v", err)
				return
			}
			slog.Debug("Recv client message %v", event)
			// todo distribute event
			// recv notify info
			selfTaskId := event.TargetOpTaskId
			targetTaskId := event.SourceOpTaskId
			bf := c.tm.bufferPool[selfTaskId]
			if bf == nil {
				// send response to upstream current download stream has closed rel task
				// stream.Send()
				stream.Send(NewAckEventReq("false", selfTaskId, targetTaskId))
			} else {
				// output data
				var credit int
				bytesBuffer := bytes.NewBuffer(event.Data)
				err = binary.Read(bytesBuffer, binary.BigEndian, &credit)
				if err != nil {
					slog.Error("binary.Read failed: %v", err)
				}
				outputDataFunc := func(data [][]byte) error {
					sendData := encodeByteData(data)
					// push data
					stream.Send(NewSingleEventReq(sendData, selfTaskId, targetTaskId))
					// ack data push result
					ack, ackErr := stream.Recv()
					if ackErr != nil || ack.EventType != pb.EventType_ACK || string(ack.Data) == "false" {
						return errors.New("data out put error")
					}
					return nil
				}
				err = bf.RemoveData(OutputData, credit, outputDataFunc)
				if err != nil {
					slog.Error("data out put error,targetTaskId:%v, selfTaskId:%v", targetTaskId, selfTaskId)
				}
			}
		}
	}()
	wg.Wait()
	return nil
}

func NewSingleEventReq(data []byte, sourceTaskId string, targetTaskId string) *pb.Event {
	return &pb.Event{
		EventTime:      time.Now().Unix(),
		Data:           data,
		SourceOpTaskId: sourceTaskId,
		TargetOpTaskId: targetTaskId,
	}
}

func NewAckEventReq(ask string, sourceTaskId string, targetTaskId string) *pb.Event {
	return &pb.Event{
		EventTime:      time.Now().Unix(),
		Data:           []byte(ask),
		SourceOpTaskId: sourceTaskId,
		TargetOpTaskId: targetTaskId,
		EventType:      pb.EventType_ACK,
	}
}

func NewCoupleEventsReq() *pb.EventChannelReq {
	return nil
}
