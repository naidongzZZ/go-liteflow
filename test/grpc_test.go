package test

import (
	"context"
	"go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"net"
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
		if err = stream.Send(&pb.EventChannelReq{Events: []*pb.Event{{Data: []byte("hello")}}}); err != nil {
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
	pb.RegisterCoreServer(srv, task_manager.NewGrpcServer())
	go func() {
		// auto stop server
		<- waitc
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
