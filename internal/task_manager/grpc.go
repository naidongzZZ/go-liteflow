package task_manager

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"io"
	"log"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// server r s r
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
				stream.Send(NewSingleEventReq([]byte("false"), selfTaskId, targetTaskId))
			} else {
				// output data
				var credit int
				bytesBuffer := bytes.NewBuffer(event.Data)
				err = binary.Read(bytesBuffer, binary.BigEndian, &credit)
				if err != nil {
					log.Fatalf("binary.Read failed: %v", err)
				}
				outputDataFunc := func(data [][]byte) error {
					sendData := encodeByteData(data)
					// push data
					stream.Send(NewSingleEventReq(sendData, selfTaskId, targetTaskId))
					// ack data push result 
					ack, _ := stream.Recv()
					if ack.EventType != pb.EventType_ACK ||  string(ack.Data) == "false" {
						return errors.New("data out put error")
					}
					return nil
				}
				err = bf.RemoveData(OutputData, credit, outputDataFunc)
				if err != nil {
					log.Fatalf("data out put error,targetTaskId:%v, selfTaskId:%v", targetTaskId, selfTaskId)
				}
			}
		}
	}()

	wg.Add(1)
	notifyc := c.tm.notifyChan
	// client , s r s
	go func(notifyc chan []any) {
		defer wg.Done()
		for notify := range notifyc {
			opId, _ := notify[0].(string)
			currentTaskId, _ := notify[1].(string)
			targetTaskId, _ := notify[2].(string)
			credit, _ := notify[3].(int16)
			stream := c.tm.GetOperatorNodeClient(opId)
			if stream != nil {
				bytesBuffer := bytes.NewBuffer([]byte{})
				binary.Write(bytesBuffer, binary.BigEndian, uint16(credit))
				// notify upstream
				err := stream.Send(NewSingleEventReq(bytesBuffer.Bytes(), currentTaskId, targetTaskId))
				if err != nil {
					break
				}
				// recv input data flow
				event, err := stream.Recv()
				if err == nil && event.EventType != pb.EventType_ACK {
					// data input
					bf := c.tm.bufferPool[currentTaskId]
					if bf != nil {
						dataFlow := decodeByteData(event.Data)
						res := bf.AddData(InputData, dataFlow)
						if !res {
							log.Fatalf("data in put error,targetTaskId:%v, selfTaskId:%v", targetTaskId, currentTaskId)
							stream.Send(NewAckEventReq("false", currentTaskId, targetTaskId))
						} else {
							stream.Send(NewAckEventReq("true", currentTaskId, targetTaskId))
						}
					} else {
						log.Fatalf("data in put error,no rel task buffer,targetTaskId:%v, selfTaskId:%v", targetTaskId, currentTaskId)
						stream.Send(NewAckEventReq("false", currentTaskId, targetTaskId))
					}
				}
			}
		}
	}(notifyc)
	
	wg.Wait()
	return nil
}

func (tm *taskManager) GetOperatorNodeClient(opId string) pb.Core_EventChannelClient {
	var client pb.Core_EventChannelClient
	var err error
	client = tm.eventChanClient[opId]
	if client == nil && tm.clientConns[opId] != nil {
		client, err = tm.clientConns[opId].EventChannel(context.Background())
		if err != nil {
			slog.Error("Failed to create event client : %v", err)
			return nil
		}
		tm.eventChanClient[opId] = client
	}
	return client
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
		EventType: pb.EventType_ACK,
	}
}


func NewCoupleEventsReq() *pb.EventChannelReq {
	return nil
}

func (c * grpcServer) DirectedEventChannel(srv pb.Core_DirectedEventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method DirectedEventChannel not implemented")
}