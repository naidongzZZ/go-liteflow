package task_manager

import (
	"context"
	"errors"
	"go-liteflow/internal/core"
	"go-liteflow/internal/pkg/operator"
	pb "go-liteflow/pb"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type taskManager struct {
	core.Comm

	taskManagerInfo *pb.ServiceInfo

	coordinatorInfo   *pb.ServiceInfo
	coordinatorClient pb.CoreClient

	srv *grpc.Server

	mux sync.Mutex
	// key: task_manager_id, val: service_addr, info
	serviceInfos map[string]*pb.ServiceInfo
	// key: service_addr, val: grpc_client
	clientConns map[string]pb.CoreClient
	// key：task_manager_id, val: RemoteEventChannel
	TaskManangerEventChans map[string]*Channel

	TaskManagerBufferMonitor

	digraphMux sync.Mutex
	// key: client_id, val: pb.disgraph
	// taskDigraph map[string]*pb.Digraph
	// key: optask_id, val: pb.OperatorTask
	tasks map[string]*pb.OperatorTask

	chMux sync.Mutex
	// key: optask_id, val: Channel
	taskChannels map[string]*Channel
}

var tm *taskManager

func GetTaskManager() *taskManager {
	return tm
}

func NewTaskManager(addr, coordAddr string) *taskManager {
	if tm != nil {
		return tm
	}
	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	tm = &taskManager{
		taskManagerInfo: &pb.ServiceInfo{
			Id:          ranUid.String(),
			ServiceAddr: addr,
			ServiceType: pb.ServiceType_TaskManager,
		},
		coordinatorInfo: &pb.ServiceInfo{
			ServiceAddr: coordAddr,
		},
		serviceInfos:             make(map[string]*pb.ServiceInfo),
		clientConns:              make(map[string]pb.CoreClient),
		tasks:                    make(map[string]*pb.OperatorTask),
		TaskManangerEventChans:   make(map[string]*Channel),
		taskChannels:             make(map[string]*Channel),
		TaskManagerBufferMonitor: *NewTaskManagerBufferMonitor(),
	}
	tm.srv = grpc.NewServer()
	pb.RegisterCoreServer(tm.srv, tm)
	return tm
}

func (tm *taskManager) RegisterTaskStreamServerClient(task *pb.OperatorTask) error {
	stream := make([]*pb.OperatorTask, 0)
	stream = append(stream, task.Downstream...)

	var wg sync.WaitGroup
	existConn := make(map[string]bool)
	for _, opTask := range stream {
		if _, e := existConn[opTask.Id]; e {
			//exist rel client
			continue
		}
		existConn[opTask.Id] = true
		wg.Add(1)
		go func() {
			addr := opTask.ClientId
			conn, _ := grpc.Dial(addr, grpc.WithInsecure())
			coreClient := pb.NewCoreClient(conn)
			tm.clientConns[addr] = coreClient
			for {
				client, cerr := coreClient.EventChannel(context.Background())
				if cerr != nil {
					// return cerr
					time.Sleep(1 * time.Second)
				} else {
					tm.eventChanClient[opTask.Id] = &client
					wg.Done()
					break
				}
			}
		}()

	}
	wg.Wait()
	// client init finish , start running task
	task.State = pb.TaskStatus_Deployed
	return nil
}


func (tm *taskManager) Start(ctx context.Context) {

	err := tm.heartBeat(ctx)
	if err != nil {
		slog.Error("task_manager heart beat.", slog.Any("err", err))
		return
	}

	tm.schedule(ctx)

	listener, err := net.Listen("tcp", tm.taskManagerInfo.ServiceAddr)
	if err != nil {
		slog.Error("new listener.", slog.Any("err", err))
		panic(err)
	}

	slog.Info("task_manager grpc server start success.")
	if err = tm.srv.Serve(listener); err != nil {
		slog.Error("grpc serve fail.", slog.Any("err", err))
		panic(err)
	}
}

func (tm *taskManager) Stop(ctx context.Context) {
	tm.srv.Stop()
}

func (tm *taskManager) Coordinator() pb.CoreClient {
	if tm.coordinatorClient != nil {
		return tm.coordinatorClient
	}

	co := tm.coordinatorInfo
	if co == nil || len(co.ServiceAddr) == 0 {
		panic("coordinator info is nil")
	}

	conn, err := grpc.Dial(co.ServiceAddr, grpc.WithInsecure())
	if err != nil {
		panic("connect coordinator fail: " + err.Error())
	}
	tm.coordinatorClient = pb.NewCoreClient(conn)
	return tm.coordinatorClient
}

func (tm *taskManager) heartBeat(ctx context.Context) (err error) {
	go func() {
		for {
			time.Sleep(5 * time.Second)

			req := &pb.HeartBeatReq{ServiceInfo: tm.taskManagerInfo}
			resp, err := tm.Coordinator().SendHeartBeat(ctx, req)
			if err != nil {
				slog.Error("send heart beat to coordinator.", slog.Any("err", err))
				continue
			}
			if resp == nil || len(resp.ServiceInfos) == 0 {
				continue
			}

			for tmid, servInfo := range resp.ServiceInfos {
				if tmid == tm.taskManagerInfo.Id {
					continue
				}
				tm.mux.Lock()
				if _, ok := tm.serviceInfos[tmid]; ok {
					// TODO update service info ?
					tm.mux.Unlock()
					continue
				}

				addr := servInfo.ServiceAddr
				conn, err := grpc.Dial(addr, grpc.WithInsecure())
				if err != nil {
					slog.Error("dial task manager", slog.String("addr", addr), slog.Any("err", err))
					tm.mux.Unlock()
					continue
				}

				client := pb.NewCoreClient(conn)
				ch, err := NewChannelWithClient(context.Background(), tmid, client)
				if err != nil {
					slog.Error("new channel.", slog.Any("err", err))
					tm.mux.Unlock()
					continue
				}

				tm.clientConns[addr] = client
				tm.TaskManangerEventChans[tmid] = ch
				tm.serviceInfos[tmid] = servInfo
				tm.mux.Unlock()
			}

			slog.Info("task manager send heart beat.", slog.Int("tm_size", len(resp.ServiceInfos)))
		}
	}()

	return
}

func (tm *taskManager) ID() string {
	return tm.taskManagerInfo.Id
}

func (tm *taskManager) GetBuffer(opId string) *Buffer {
	return tm.bufferPool[opId]
}

func (tm *taskManager) GetOperatorNodeClient(opId string) pb.Core_EventChannelClient {
	var client pb.Core_EventChannelClient
	var err error
	client = *tm.eventChanClient[opId]
	if client == nil && tm.clientConns[opId] != nil {
		client, err = tm.clientConns[opId].EventChannel(context.Background())
		if err != nil {
			slog.Error("Failed to create event client : %v", err)
			return nil
		}
		tm.eventChanClient[opId] = &client
	}
	return client
}

func (tm *taskManager) Invoke(ctx context.Context, opTask *pb.OperatorTask, ch *Channel) (err error) {

	opFn, ok := operator.GetOpFn(opTask.ClientId, opTask.OpId, opTask.OpType)
	if !ok {
		return errors.New("unsupported Operator Func")
	}

	// wait for channel ready
	dsch := make(map[string]*Channel)
	var times, maxTryTimes = 0, 30
	for len(opTask.Downstream) != len(dsch) {
		for _, ds := range opTask.Downstream {
			ch := tm.GetChannel(opTask.ClientId, ds.Id)
			if ch != nil {
				dsch[ds.Id] = ch
			}
		}
		if len(opTask.Downstream) == len(dsch) {
			break
		}
		if times > maxTryTimes {
			slog.Error("downstream channel not ready", slog.Any("dsch", dsch))
			return errors.New("channel not ready")
		}
		time.Sleep(3 * time.Second)
	}

	for {
		select {
		case ev := <-ch.InputCh():
			if ev.EventType == pb.EventType_EtUnknown {
				continue
			}

			slog.Info("operator input.", slog.Any("opTaskId", opTask.Id), slog.Any("event", ev))
			// invoke opFn
			output := opFn(ctx, ev)

			if len(opTask.Downstream) == 0 {
				continue
			}
			slog.Info("operator output.", slog.Any("opTaskId", opTask.Id), slog.Any("events", output))

			for _, oev := range output {
				// TODO distribute target opTaskId
				dsId := opTask.Downstream[0].Id
				ds, ok := dsch[dsId]
				if !ok {
					slog.Error("not found ds channel.", slog.String("optaskId", dsId))
					continue
				}

				ds.InputCh() <- &pb.Event{
					Id:             uuid.NewString(),
					EventType:      pb.EventType_DataOutPut,
					EventTime:      ev.EventTime,
					SourceOpTaskId: opTask.Id,
					TargetOpTaskId: dsId,
					Key:            oev.Key,
					Data:           oev.Data}
			}
		case <-ctx.Done():
			slog.Info("operator done.", slog.Any("opTaskId", opTask.Id), slog.Any("err", ctx.Err()))
			return
		}
	}
}

func (tm *taskManager) schedule(ctx context.Context) {
	// TODO Currently, task scheduling is not supported. start task by ManageOpTask method

	// schedule ready optask
	// new channel
	// TODO use goroutine pool
	// run goroutine
	// TODO notify optask status
}

func (tm *taskManager) RegisterChannel(channels ...*Channel) {
	if len(channels) == 0 {
		return
	}
	for _, ch := range channels {
		if ch.opTaskId != "" {
			tm.chMux.Lock()
			tm.taskChannels[ch.opTaskId] = ch
			tm.chMux.Unlock()
			continue
		}
		if ch.taskManagerId != "" {
			tm.mux.Lock()
			tm.TaskManangerEventChans[ch.taskManagerId] = ch
			tm.mux.Unlock()
			continue
		}
	}
}

func (tm *taskManager) GetChannel(clientId, optaskId string) (ch *Channel) {
	tm.chMux.Lock()
	ch = tm.taskChannels[optaskId]
	tm.chMux.Unlock()
	if ch != nil {
		return
	}

	req := &pb.FindOpTaskReq{ClientId: clientId, OpTaskId: optaskId}
	resp, err := tm.Coordinator().FindOpTask(context.TODO(), req)
	if err != nil {
		slog.Error("find opTask.", slog.Any("err", err))
		return nil
	}
	if resp == nil || resp.TaskManagerId == "" {
		return nil
	}
	tm.mux.Lock()
	defer tm.mux.Unlock()
	ch = tm.TaskManangerEventChans[resp.TaskManagerId]
	return
}
