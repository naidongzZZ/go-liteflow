package pkg

import(
	pb "go-liteflow/pb"
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