package coordinator

import (
	"context"
	"go-liteflow/internal/pkg"
	pb "go-liteflow/pb"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Coordinator 接收提交任务请求, 生成任务图
func (co *coordinator) SubmitOpTask(ctx context.Context, req *pb.SubmitOpTaskReq) (resp *pb.SubmitOpTaskResp, err error) {
	resp = new(pb.SubmitOpTaskResp)

	slog.Debug("Recv submit op task.", slog.Any("req", req))

	digraph := req.Digraph

	if len(digraph.Adj) != 1 {
		return resp, status.Errorf(codes.InvalidArgument, "")
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
		return resp, status.Errorf(codes.InvalidArgument, "")
	}

	co.digraphMux.Lock()
	defer co.digraphMux.Unlock()
	co.taskDigraph[clientId] = di

	return resp, nil
}
