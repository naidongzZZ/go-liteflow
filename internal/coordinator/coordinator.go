package coordinator

import (
	"context"
	"net"
	"sync"
	"time"

	"go-liteflow/internal/core"
	pb "go-liteflow/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type coordinator struct {
	core.Comm

	// 服务id
	Id string 
	// mux
	mux sync.Mutex
	// key: 服务id
	serviceInfos map[string]*core.Service

	srv *grpc.Server
	
	// key: client_id, val: pb.disgraph
	digraphMux  sync.Mutex
	taskDigraph map[string]*pb.Digraph
}

func NewCoordinator(addr string) *coordinator {

	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	co := &coordinator{
		Id:          ranUid.String(),
		serviceInfos: make(map[string]*core.Service),
		taskDigraph:  make(map[string]*pb.Digraph),
	}

	// 将自身信息注册到serviceInfos
	co.serviceInfos[co.Id] = &core.Service{
		ServiceInfo: pb.ServiceInfo{
			Id:            co.Id,
			ServiceAddr:   addr,
			ServiceType:   pb.ServiceType_Coordinator,
			ServiceStatus: pb.ServiceStatus_SsRunning,
			Timestamp:     time.Now().Unix()},
	}

	co.srv = grpc.NewServer()
	pb.RegisterCoreServer(co.srv, co)

	return co
}

func (co *coordinator) Start(ctx context.Context) {

	co.schedule(ctx)

	listener, err := net.Listen("tcp", co.SelfServiceInfo().ServiceAddr)
	if err != nil {
		panic(err)
	}
	if err = co.srv.Serve(listener); err != nil {
		panic(err)
	}
}

func (co *coordinator) ID() string {
	return co.Id
}

func (co *coordinator) SelfServiceInfo() *core.Service {
	return co.serviceInfos[co.Id]
}
