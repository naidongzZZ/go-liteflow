package task_manager

import (
	"context"
	"go-liteflow/internal/core"
	"go-liteflow/internal/pkg/log"
	"go-liteflow/internal/pkg/storager"
	pb "go-liteflow/pb"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type taskManager struct {
	core.CoreComm

	// 服务id
	Id string
	// coordinator的id
	CoId string

	srv *grpc.Server

	mux sync.Mutex
	// key: 服务id
	serviceInfos map[string]*core.Service

	digraphMux sync.Mutex
	// key: optask_id, val: *pb.Digraph
	tasks map[string]*pb.Digraph

	chMux sync.Mutex
	// key: optask_id, val: Channel
	taskConn map[string]net.Conn

	storager storager.Storager
}

var tm *taskManager

func GetTaskManager() *taskManager {
	return tm
}

func NewTaskManager(addr, coordAddr string) *taskManager {
	if tm != nil {
		return tm
	}
	if len(coordAddr) == 0 {
		panic("coordinator's address is nil")
	}

	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	tm = &taskManager{
		Id:           ranUid.String(),
		serviceInfos: make(map[string]*core.Service),
		tasks:        make(map[string]*pb.Digraph),
		storager:     storager.NewStorager(context.Background(), "/tmp/task_ef"),
		taskConn:     make(map[string]net.Conn),
	}

	tm.serviceInfos[tm.Id] = &core.Service{
		ServiceInfo: pb.ServiceInfo{
			Id:            tm.Id,
			ServiceAddr:   addr,
			ServiceType:   pb.ServiceType_TaskManager,
			ServiceStatus: pb.ServiceStatus_SsRunning,
			Timestamp:     time.Now().Unix()},
	}

	coCli, err := core.NewCoreClient(coordAddr)
	if err != nil {
		slog.Error("new coordinator client.", slog.Any("err", err))
		panic(err)
	}
	tm.serviceInfos[tm.CoId] = &core.Service{
		ServiceInfo: pb.ServiceInfo{
			Id:            tm.CoId,
			ServiceAddr:   coordAddr,
			ServiceType:   pb.ServiceType_Coordinator,
			ServiceStatus: pb.ServiceStatus_SsRunning,
			Timestamp:     time.Now().Unix()},
		ClientConn: coCli,
	}

	tm.srv = grpc.NewServer(grpc.MaxRecvMsgSize(1024 * 1024 * 1024))
	pb.RegisterCoreServer(tm.srv, tm)
	return tm
}

func (tm *taskManager) Start(ctx context.Context) {

	err := tm.heartBeat(ctx)
	if err != nil {
		slog.Error("task_manager heart beat.", slog.Any("err", err))
		return
	}

	tm.schedule(ctx)

	GracefulRun(func(c context.Context) error {
		addr := tm.SelfServiceInfo().ServiceAddr
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		log.Infof("task_manager grpc server start success")

		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					tm.srv.Stop()
					return
				}
			}
		}(c)

		if err = tm.srv.Serve(lis); err != nil {
			panic(err)
		}
		log.Infof("task_manager stop grpc server")
		return nil
	})

	GracefulRun(func(c context.Context) error {
		lis, err := net.Listen("unix", UnixAddress(tm.Id))
		if err != nil {
			panic(err)
		}

		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					lis.Close()
					return
				}
			}
		}(c)

		ServeOn(c, lis)
		log.Infof("task_manager stop unix socket")
		return nil
	})

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGQUIT, syscall.SIGHUP)
	<-ch
	GracefulStop()
}

func (tm *taskManager) Stop(ctx context.Context) {
	tm.srv.Stop()
}

func (tm *taskManager) Coordinator() pb.CoreClient {
	return tm.serviceInfos[tm.CoId].ClientConn
}

func (tm *taskManager) ID() string {
	return tm.Id
}

func (tm *taskManager) SelfServiceInfo() *core.Service {
	return tm.serviceInfos[tm.Id]
}

func (tm *taskManager) schedule(ctx context.Context) {
	// TODO Currently, task scheduling is not supported. start task by ManageOpTask method

	// schedule ready optask
	// new channel
	// TODO use goroutine pool
	// run goroutine
	// TODO notify optask status
}

func (tm *taskManager) GetTaskConn(taskId string) (net.Conn, bool) {
	tm.chMux.Lock()
	defer tm.chMux.Unlock()
	conn, ok := tm.taskConn[taskId]
	return conn, ok
}

func (tm *taskManager) SetTaskConn(taskId string, conn net.Conn) {
	tm.chMux.Lock()
	defer tm.chMux.Unlock()
	_, ok := tm.taskConn[taskId]
	if !ok {
		tm.taskConn[taskId] = conn
	}
}
