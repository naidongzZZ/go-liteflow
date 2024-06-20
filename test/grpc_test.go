package test

import (
	"context"
	"go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"log/slog"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestEventChannelUpstream(t *testing.T) {
	taskId := "upstreamTask"
	tm1 := task_manager.NewTaskManager("127.0.0.1:20020", "")
	t1 := &pb.OperatorTask{
		Id:         taskId,
		OpType:     pb.OpType_Map,
		Downstream: make([]*pb.OperatorTask, 0),
	}

	t1.Downstream = append(t1.Downstream, &pb.OperatorTask{Id: "insidestreamTask", ClientId: "127.0.0.1:20022", TaskManagerId: "127.0.0.1:20022"})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		go func() {
			time.Sleep(1 * time.Minute)
			tm1.CloseOperatorTask(taskId)
			tm1.Stop(context.Background())
		}()
		tm1.Start(context.Background())
	}()
	tm1.RegisterOperatorTask(t1)
	for i := 0; i < 5; i++ {
		go tm1.GetBuffer(t1.Id).AddData(task_manager.OutputData, []*pb.Event{{
			Data:           []byte("hello" + strconv.Itoa(i)),
			TargetOpTaskId: "insidestreamTask",
		}})
	}
	wg.Wait()
}

func TestEventChannelInsideStream(t *testing.T) {
	taskId := "insidestreamTask"
	tm2 := task_manager.NewTaskManager("127.0.0.1:20022", "")
	t2 := &pb.OperatorTask{
		Id:       taskId,
		OpType:   pb.OpType_Map,
		Upstream: make([]*pb.OperatorTask, 0),
	}
	t2.Upstream = append(t2.Upstream, &pb.OperatorTask{Id: "upstreamTask", ClientId: "127.0.0.1:20020", TaskManagerId: "127.0.0.1:20020"})
	t2.Downstream = append(t2.Downstream, &pb.OperatorTask{Id: "downstreamTask", ClientId: "127.0.0.1:20021", TaskManagerId: "127.0.0.1:20021"})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		go func() {
			time.Sleep(1 * time.Minute)
			tm2.CloseOperatorTask(taskId)
			tm2.Stop(context.Background())
		}()
		tm2.Start(context.Background())
	}()
	tm2.RegisterOperatorTask(t2)
	//read input queue data and handler them to output queue
	tm2.GetBuffer(taskId).ListenInputQueueAndHandle(func(data *pb.Event) error {
		slog.Info("handle data:", slog.Any("event data", data))
		tm2.GetBuffer(t2.Id).AddData(task_manager.OutputData, []*pb.Event{{
			Data:           []byte("insight data" + regexp.MustCompile(`\d`).FindString(string(data.Data))),
			TargetOpTaskId: "downstreamTask",
		}})
		return nil
	})
	wg.Wait()
}

func TestEventChannelDownstream(t *testing.T) {
	taskId := "downstreamTask"
	tm2 := task_manager.NewTaskManager("127.0.0.1:20021", "")
	t2 := &pb.OperatorTask{
		Id:       taskId,
		OpType:   pb.OpType_Map,
		Upstream: make([]*pb.OperatorTask, 0),
	}
	t2.Upstream = append(t2.Upstream, &pb.OperatorTask{Id: "upstreamTask", ClientId: "127.0.0.1:20022", TaskManagerId: "127.0.0.1:20022"})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		go func() {
			time.Sleep(1 * time.Minute)
			tm2.CloseOperatorTask(taskId)
			tm2.Stop(context.Background())
		}()
		tm2.Start(context.Background())
	}()
	tm2.RegisterOperatorTask(t2)
	//read input queue data and handler them to output queue
	tm2.GetBuffer(taskId).ListenInputQueueAndHandle(func(data *pb.Event) error {
		slog.Info("handle data:", slog.Any("event data", data))
		time.Sleep(5 * time.Second)
		return nil
	})
	wg.Wait()
}
