package test

import (
	"context"
	"go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestEventChannel(t *testing.T) {

	waitc := make(chan struct{}, 1)

	// client
	go func() {
		// wait for the server to start
		time.Sleep(2 * time.Second)

		conn, err := grpc.Dial("127.0.0.1:20020", grpc.WithInsecure())
		if err != nil {
			t.Log(err)
			return
		}
		stream, err := pb.NewCoreClient(conn).EventChannel(context.Background())
		if err != nil {
			t.Log(err)
			return
		}
		if err = stream.Send(&pb.Event{Data: []byte("hello")}); err != nil {
			t.Log(err)
			return
		}
		stream.CloseSend()

		eventChannelResp, err := stream.Recv()
		if err != nil {
			t.Log(err)
			return
		}
		t.Log("recv server resp: ", eventChannelResp)
		waitc <- struct{}{}
	}()

	// start server
	srv := grpc.NewServer()
	pb.RegisterCoreServer(srv, task_manager.NewTaskManager("", ""))
	go func() {
		// auto stop server
		<-waitc
		t.Log("closing server")
		srv.Stop()
	}()

	listener, err := net.Listen("tcp", "127.0.0.1:20020")
	if err != nil {
		t.Fatal(err)
	}
	if err = srv.Serve(listener); err != nil {
		t.Fatal(err)
	}

}

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
		tm1.Start(context.Background())
	}()
	tm1.RegisterOperatorTask(t1)
	tm1.RegisterTaskStreamServerClient(t1)
	time.Sleep(2 * time.Second)
	for i := 0; i < 5; i++ {
		tm1.GetBuffer(t1.Id).AddData(task_manager.OutputData, []*pb.Event{{
			Data:           []byte("hello" + strconv.Itoa(i)),
			TargetOpTaskId: "insidestreamTask",
		}})
		time.Sleep(5 * time.Second)
	}
	tm1.CloseOperatorTask(taskId)
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
		tm2.Start(context.Background())
	}()
	tm2.RegisterOperatorTask(t2)
	tm2.RegisterTaskStreamServerClient(t2)
	in := tm2.GetBuffer(taskId).InQueue
	for data := range in {
		tm2.GetBuffer(t2.Id).AddData(task_manager.OutputData, []*pb.Event{{
			Data:           data.Data,
			TargetOpTaskId: "downstreamTask",
		}})
		time.Sleep(5 * time.Second)
	}
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
		tm2.Start(context.Background())
	}()
	tm2.RegisterOperatorTask(t2)
	tm2.RegisterTaskStreamServerClient(t2)

	wg.Wait()
}

func TestCoordinatorSubmitOpTask(t *testing.T) {

}
