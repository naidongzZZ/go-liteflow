package task_manager

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	pb "go-liteflow/pb"
	"io"
	"log"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (tm *taskManager) EventChannel(stream pb.Core_EventChannelServer) error {

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
			bf := tm.bufferPool[selfTaskId]
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
	wg.Add(1)
	notifyc := tm.notifyChan
	// client , s r s
	go func(notifyc chan NotifyEvent) {
		defer wg.Done()
		for notify := range notifyc {
			opId := notify.opId
			currentTaskId := notify.sourceTaskId
			targetTaskId := notify.targetTaskId
			credit := notify.credit
			stream := tm.GetOperatorNodeClient(opId)
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
					bf := tm.bufferPool[currentTaskId]
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

func (tm *taskManager) DeployOpTask(ctx context.Context, req *pb.DeployOpTaskReq) (resp *pb.DeployOpTaskResp, err error) {
	resp = new(pb.DeployOpTaskResp)
	tm.digraphMux.Lock()
	defer tm.digraphMux.Unlock()

	tm.taskDigraph[req.ClientId] = req.Digraph

	for _, optaskId := range req.OpTaskIds {
		for i := range req.Digraph.Adj {
			if optaskId == req.Digraph.Adj[i].Id {
				slog.Debug("Deploy Optask:%s to TaskManager:%s", optaskId, tm.ID())
				tm.tasks[optaskId] = req.Digraph.Adj[i]
			}
		}
	}

	return resp, nil
}
func (tm *taskManager) ManageOpTask(ctx context.Context, req *pb.ManageOpTaskReq) (resp *pb.ManageOpTaskResp, err error) {
	return nil, status.Errorf(codes.Unimplemented, "method ManageOpTask not implemented")
}

func (tm *taskManager) DirectedEventChannel(srv pb.Core_DirectedEventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method DirectedEventChannel not implemented")
}
