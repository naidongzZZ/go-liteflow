package coordinator

import (
	"context"
	"net"

	pb "go-liteflow/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type coordinator struct {
	coordinatorId string
	addr          string

	srv  *grpc.Server
	gSrv *grpcServer

	taskManagerAddrs map[string]struct{}
}

func NewCoordinator(addr string) *coordinator {

	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	co := &coordinator{
		coordinatorId:    ranUid.String(),
		addr:             addr,
		taskManagerAddrs: map[string]struct{}{},
	}

	srv := grpc.NewServer()

	gSrv := &grpcServer{coord: co}
	pb.RegisterCoreServer(srv, gSrv)

	co.gSrv = gSrv
	co.srv = srv

	return co
}

func (co *coordinator) Start(ctx context.Context) {

	listener, err := net.Listen("tcp", co.addr)
	if err != nil {
		panic(err)
	}
	if err = co.srv.Serve(listener); err != nil {
		panic(err)
	}
}

func (co *coordinator) ID() string {
	return co.coordinatorId
}
