package task_manager

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"go-liteflow/internal/pkg"
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
			// TODO distribute event
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

	opTaskIds := make([]string, 0)
	for i := range req.Digraph.Adj {
		// validate opTask
		opTask := req.Digraph.Adj[i]

		if e := pkg.ValidateOpTask(opTask); e != nil {
			slog.Warn("validate optask.", slog.Any("err", e))
			return resp, status.Error(codes.InvalidArgument, "")
		}

		// TODO download plugin.so

		slog.Debug("Deploy Optask:%s to TaskManager:%s", opTask.Id, tm.ID())
		tm.tasks[opTask.Id] = req.Digraph.Adj[i]
		opTaskIds = append(opTaskIds, opTask.Id)
	}

	// TODO deploy and then start task
	_, err = tm.ManageOpTask(ctx, &pb.ManageOpTaskReq{ManageType: pb.ManageType_Start, OpTaskIds: opTaskIds})
	if err != nil {
		slog.Error("start optask.", slog.Any("opTaskIds", opTaskIds), slog.Any("err", err))
		return resp, err
	}

	return resp, nil
}

func (tm *taskManager) ManageOpTask(ctx context.Context, req *pb.ManageOpTaskReq) (resp *pb.ManageOpTaskResp, err error) {

	resp = new(pb.ManageOpTaskResp)

	switch req.ManageType {
	case pb.ManageType_Start:
		{
			tm.digraphMux.Lock()
			defer tm.digraphMux.Unlock()

			for _, opTaskId := range req.OpTaskIds {
				task, ok := tm.tasks[opTaskId]
				// TODO task state transition
				if ok && task.State == pb.TaskStatus_Ready {
					tm.chMux.Lock()
					ch, ok := tm.taskChannels[opTaskId]
					if !ok {
						ch = NewChannel(opTaskId)
						tm.taskChannels[opTaskId] = ch
					}
					tm.chMux.Unlock()

					// TODO WithTimeout context 
					go tm.Invoke(context.TODO(), task, ch)

					// TODO notify optask status to coordinator
					task.State = pb.TaskStatus_Running
				}
			}
		}
	default:
		return resp, status.Errorf(codes.Unimplemented, "manage type %v unsupported", req.ManageType)
	}

	return
}

func (tm *taskManager) DirectedEventChannel(srv pb.Core_DirectedEventChannelServer) (err error) {

	for {
		eventChannelReq, e := srv.Recv()
		if e == io.EOF {
			return
		}
		if e != nil {
			slog.Error("directed channel recv.", slog.Any("err", e))
			return
		}
		if eventChannelReq == nil || len(eventChannelReq.Events) == 0 {
			continue
		}

		tm.chMux.Lock()
		for _, ev := range eventChannelReq.Events {
			ch, ok := tm.taskChannels[ev.TargetOpTaskId]
			if !ok || ch == nil {
				// TODO ack ?
				continue
			}

			subCtx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
			ch.PutInputCh(subCtx, ev)
			cancel()
		}
		tm.chMux.Unlock()

	}
	return err
}
