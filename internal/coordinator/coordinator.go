package coordinator

import (
	"context"
	"errors"
	"net"
	"sync"

	"go-liteflow/internal/pkg"
	pb "go-liteflow/pb"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
)

type coordinator struct {
	coordinatorInfo *pb.ServiceInfo

	srv  *grpc.Server
	gSrv *grpcServer

	mux sync.Mutex
	// key: coordinator_id or task_manager_id
	serviceInfos map[string]*pb.ServiceInfo

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
		coordinatorInfo: &pb.ServiceInfo{
			Id:          ranUid.String(),
			ServiceAddr: addr,
			ServiceType: pb.ServiceType_Coordinator,
		},
		serviceInfos: make(map[string]*pb.ServiceInfo),
		taskDigraph:  make(map[string]*pb.Digraph),
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

func (co *coordinator) SubmitOpTask(digraph *pb.Digraph) (err error) {

	if len(digraph.Adj) != 1 {
		return errors.New("illegal OpTask")
	}

	tmpLevelTask := make([][]*pb.OperatorTask, 0)

	clientId := ""

	// generate op task instance by parallelism
	cur := digraph.Adj[0]
	for cur != nil {

		if clientId == "" {
			clientId = cur.ClientId
		}

		curLevel := make([]*pb.OperatorTask, 0, int(cur.Parallelism))
		for i := 0; i < int(cur.Parallelism); i++ {
			curLevel[i] = new(pb.OperatorTask)
			copier.Copy(curLevel[i], cur)

			curLevel[i].Id = pkg.OpTaskId(curLevel[i].OpType, i)

			curLevel[i].Upstream = make([]*pb.OperatorTask, 0)
			curLevel[i].Downstream = make([]*pb.OperatorTask, 0)
		}
		tmpLevelTask = append(tmpLevelTask, curLevel)

		if len(cur.Downstream) == 0 {
			cur = nil
		} else {
			cur = cur.Downstream[0]
		}
	}

	for i := 0; i < len(tmpLevelTask); i++ {

		// TODO more patterns
		curLevelTaskIds := make([]string, 0, len(tmpLevelTask[i]))
		curLevelTask := make(map[string]*pb.OperatorTask)
		for j := range tmpLevelTask[i] {
			t := tmpLevelTask[i][j]
			curLevelTaskIds = append(curLevelTaskIds, t.Id)
			curLevelTask[t.Id] = t
		}

		nextLevelTaskIds := make([]string, 0)
		nextLevelTask := make(map[string]*pb.OperatorTask)

		if i+1 < len(tmpLevelTask) && len(tmpLevelTask[i+1]) > 0 {
			for j := range tmpLevelTask[i+1] {
				t := tmpLevelTask[i+1][j]
				nextLevelTaskIds = append(nextLevelTaskIds, t.Id)
				nextLevelTask[t.Id] = t
			}
		}

		for _, nextLvTid := range nextLevelTaskIds {

			if nextLvTask, ok := nextLevelTask[nextLvTid]; ok {
				upstreamTasks := pkg.ToOpTasks(curLevelTaskIds,
					func(s string) *pb.OperatorTask { return &pb.OperatorTask{Id: s} })
				nextLvTask.Upstream = append(nextLvTask.Upstream, upstreamTasks...)
			}

			for _, curLvTid := range curLevelTaskIds {
				if curLvTask, ok := curLevelTask[curLvTid]; ok {
					downstreamTask := &pb.OperatorTask{Id: nextLvTid}
					curLvTask.Downstream = append(curLvTask.Downstream, downstreamTask)
				}
			}
		}
	}

	di := &pb.Digraph{GraphId: uuid.NewString(), Adj: tmpLevelTask[0]}
	if err = uuid.Validate(clientId); err != nil {
		return err
	}

	co.digraphMux.Lock()
	defer co.digraphMux.Unlock()
	co.taskDigraph[clientId] = di

	return nil
}

func (co *coordinator) schedule() {
}
