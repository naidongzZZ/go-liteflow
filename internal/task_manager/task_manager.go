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

	coordinatorInfo *pb.ServiceInfo

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
	taskDigraph map[string]*pb.Digraph
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
		serviceInfos:           make(map[string]*pb.ServiceInfo),
		clientConns:            make(map[string]pb.CoreClient),
		taskDigraph:            make(map[string]*pb.Digraph),
		tasks:                  make(map[string]*pb.OperatorTask),
		TaskManangerEventChans: make(map[string]*Channel),
		taskChannels:           make(map[string]*Channel),
	}
	tm.srv = grpc.NewServer()
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

func (tm *taskManager) heartBeat(ctx context.Context) (err error) {
	conn, err := grpc.Dial(tm.coordinatorInfo.ServiceAddr, grpc.WithInsecure())
	if err != nil {
		slog.Error("dial coordinator", slog.Any("err", err))
		return err
	}

	coordinatorConn := pb.NewCoreClient(conn)
	tm.mux.Lock()
	tm.clientConns[tm.taskManagerInfo.ServiceAddr] = coordinatorConn
	tm.mux.Unlock()

	go func() {
		for {
			time.Sleep(5 * time.Second)

			req := &pb.HeartBeatReq{ServiceInfo: tm.taskManagerInfo}
			resp, err := coordinatorConn.SendHeartBeat(ctx, req)
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

func (tm *taskManager) GetOperatorNodeClient(opId string) pb.Core_EventChannelClient {
	var client pb.Core_EventChannelClient
	var err error
	client = tm.eventChanClient[opId]
	if client == nil && tm.clientConns[opId] != nil {
		client, err = tm.clientConns[opId].EventChannel(context.Background())
		if err != nil {
			slog.Error("Failed to create event client : %v", err)
			return nil
		}
		tm.eventChanClient[opId] = client
	}
	return client
}

func (tm *taskManager) Invoke(ctx context.Context, opTask *pb.OperatorTask, ch *Channel) (err error) {

	opFn, ok := operator.GetOpFn(opTask.ClientId, opTask.OpId, opTask.OpType)
	if !ok {
		return errors.New("unsupported Operator Func")
	}

	downstreamCh := make(map[string]*Channel)
	tm.chMux.Lock()
	for _, ds := range opTask.Downstream {
		tmp, ok := tm.taskChannels[ds.Id]
		if !ok {
			tm.mux.Lock()
			// TODO not found task manager
			tmp, ok = tm.TaskManangerEventChans[ds.TaskManagerId]
			if !ok {
				slog.Info("not found channel.", slog.String("downstream", ds.Id))
				tm.mux.Unlock()
				tm.chMux.Unlock()
				return
			}
			tm.mux.Unlock()
		}
		downstreamCh[ds.Id] = tmp
	}
	tm.chMux.Unlock()

	for {
		select {
		case ev := <-ch.InputCh():
			if ev.EventType == pb.EventType_EtUnknown {
				return
			}

			slog.Info("operator input.", slog.Any("opTaskId", opTask.Id), slog.Any("event", ev))

			output := opFn(ctx, ev)

			if len(opTask.Downstream) != 0 {
				slog.Info("operator output.", slog.Any("opTaskId", opTask.Id), slog.Any("events", output))
				// TODO distribute target opTaskId

				for _, oev := range output {
					// TODO which opTaskId ?
					dsId := opTask.Downstream[0].Id
					downstream, ok := downstreamCh[dsId]
					if !ok {
						slog.Error("not found downstream.", slog.String("optaskId", dsId))
						return
					}

					downstream.InputCh() <- &pb.Event{
						Id:             uuid.NewString(),
						EventType:      pb.EventType_DataOutPut,
						EventTime:      ev.EventTime,
						SourceOpTaskId: opTask.Id,
						TargetOpTaskId: dsId,
						Key:            oev.Key,
						Data:           oev.Data}
				}
			}

		case <-ctx.Done():
			slog.Info("operator done.", slog.Any("opTaskId", opTask.Id), slog.Any("err", ctx.Err()))
			return
		}
	}
}

func (tm *taskManager) schedule(ctx context.Context) {
	go func() {

		for {
			time.Sleep(1 * time.Second)

			var curTask *pb.OperatorTask

			// schedule ready optask
			tm.digraphMux.Lock()
			for _, task := range tm.tasks {
				if task.State == pb.TaskStatus_Ready {
					curTask = task
					break
				}
			}
			tm.digraphMux.Unlock()

			if curTask == nil {
				continue
			}
			// new channel
			tm.chMux.Lock()
			ch, ok := tm.taskChannels[curTask.Id]
			if !ok {
				ch = NewChannel(curTask.Id)
				tm.taskChannels[curTask.Id] = ch
			}
			tm.chMux.Unlock()

			// TODO use goroutine pool
			// run goroutine
			go tm.Invoke(context.Background(), curTask, ch)

			// TODO notify optask status
		}
	}()
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
