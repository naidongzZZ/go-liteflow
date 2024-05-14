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

	ch := make(chan struct{}, 1)

	go func() {
		<-ch
		conn, err := grpc.Dial("127.0.0.1:20020", grpc.WithInsecure())
		if err != nil {
			panic(err)
		}

		stream, err := pb.NewCoreClient(conn).EventChannel(context.Background())
		if err != nil {
			panic(err)
		}
		if err = stream.Send(&pb.EventChannelReq{Events: []*pb.Event{{Data: []byte("hello")}}}); err != nil {
			panic(err)
		}
		stream.CloseSend()
	}()

	go func() {
		time.Sleep(1 * time.Second)
		ch <- struct{}{}
	}()

	srv := grpc.NewServer()
	pb.RegisterCoreServer(srv, task_manager.NewGrpcServer())
	go func() {
		time.Sleep(2 * time.Second)
		srv.GracefulStop()
	}()

	listener, err := net.Listen("tcp", "127.0.0.1:20020")
	if err != nil {
		t.Fatal(err)
	}
	if err = srv.Serve(listener); err != nil {
		t.Fatal(err)
	}

}
