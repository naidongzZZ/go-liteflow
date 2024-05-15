package coordinator

import (
	"context"
	"errors"
	"net"
	"sync"

	"go-liteflow/internal/pkg"
	pb "go-liteflow/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type coordinator struct {
	coordinatorInfo *pb.ServiceInfo

	srv  *grpc.Server
	gSrv *grpcServer

	mux          sync.Mutex
	serviceInfos map[string]*pb.ServiceInfo
}

func NewCoordinator(addr string) *coordinator {

	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	co := &coordinator{
		coordinatorInfo: &pb.ServiceInfo{
			Id:          ranUid.String(),
			ServiceAddr: addr,
			ServiceType: pb.ServiceType_Coordinator,
		},
		serviceInfos: make(map[string]*pb.ServiceInfo),
	}

	srv := grpc.NewServer()

	gSrv := &grpcServer{coord: co}
	pb.RegisterCoreServer(srv, gSrv)

	co.gSrv = gSrv
	co.srv = srv

	return co
}

func (co *coordinator) Start(ctx context.Context) {

	listener, err := net.Listen("tcp",
		co.coordinatorInfo.ServiceAddr)
	if err != nil {
		panic(err)
	}
	if err = co.srv.Serve(listener); err != nil {
		panic(err)
	}
}

func (co *coordinator) ID() string {
	return co.coordinatorInfo.Id
}

func (co *coordinator) RegistServiceInfo(si *pb.ServiceInfo) (err error) {
	if err = uuid.Validate(si.Id); err != nil {
		return err
	}
	if !pkg.ValidateIPv4WithPort(si.ServiceAddr) {
		return errors.New("addr is illegal")
	}
	co.mux.Lock()
	defer co.mux.Unlock()

	co.serviceInfos[si.Id] = si
	return nil
}

func (co *coordinator) GetServiceInfo(ids ...string) map[string]*pb.ServiceInfo {
	co.mux.Lock()
	defer co.mux.Unlock()

	if len(ids) == 0 {
		return co.serviceInfos
	}
	tmp := make(map[string]*pb.ServiceInfo)
	for _, id := range ids {
		tmp[id] = co.serviceInfos[id]
	}
	return tmp
}
