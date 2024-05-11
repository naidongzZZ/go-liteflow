package task_manager

import (
	pb "go-liteflow/pb"
	"net"

	"google.golang.org/grpc"
)

type taskManager struct {
	TaskManagerId string
	Addr string

	CoordinatorId   string
	CoordinatorAddr string


}

func NewTaskManager(addr, coordAddr string) *taskManager {

	tm := &taskManager{}

	srv := grpc.NewServer()

	gSrv := &grpcServer{tm: tm}
	pb.RegisterCoreServer(srv, gSrv)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	go func ()  {
		if err = srv.Serve(listener); err != nil {
			panic(err)
		}
	} ()

	return tm
}
