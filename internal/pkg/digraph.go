package pkg

import (
	"fmt"
	pb "go-liteflow/pb"

	"github.com/google/uuid"
)

// todo directed graph util

func HasCycle(g *pb.Digraph) bool {
	
	return false
}

func FindOpTask(g *pb.Digraph, optaskId string) (t *pb.OperatorTask) {
	for i, optask := range g.Adj {
		if optask.Id == optaskId {
			return g.Adj[i]
		}
	}
	return nil
}

func FindDownstreamOpTask(g *pb.Digraph, optaskId string) (t []*pb.OperatorTask) {
	cur := FindOpTask(g, optaskId)
	if cur == nil {
		return nil
	}
	return cur.Downstream
}

func OpTaskId(opType pb.OpType, seq int) string {
	return fmt.Sprintf("%s%d-%s", opType.String(), seq, uuid.NewString())
}

func ToOpTasks[T any](source []string, fn func(string) T) []T {
	res := make([]T, 0, len(source))
	for i, e := range source {
		res[i] = fn(e)
	}
	return res
}

