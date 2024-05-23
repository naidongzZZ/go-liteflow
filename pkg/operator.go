package pkg

import (
	pb "go-liteflow/pb"
)

type MapOp interface {

	Map([]*pb.Event) []*pb.Event

}

type ReduceOp interface {
	// todo 
	Reduce(map[string]int, []*pb.Event) []*pb.Event

}