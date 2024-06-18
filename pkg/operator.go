package pkg

import (
	"fmt"
	tm "go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"hash/fnv"
	"time"
)

type Operation struct {
	Type       pb.OpType
	Func       func(Output func([]byte, ...any))
	OutputFunc func([]byte, ...any)
	Ot         *pb.OperatorTask
}

func (o *Operation) LocalEnvOutput(data []byte, unuseParam ...any) {
	for _, d := range data {
		fmt.Println(string(d))
	}

}

func (o *Operation) LocalEnvKeyByOutput(data []byte, key ...any) {
	nextOperaParallelism := uint(len(o.Ot.Downstream))
	slot := getSlot(anyToBytes(key), nextOperaParallelism)
	ds := o.Ot.Downstream[slot]
	event := &pb.Event{
		TargetOpTaskId: ds.Id,
		SourceOpTaskId: o.Ot.Id,
		Data:           data,
		EventType:      pb.EventType_DataOutPut,
		EventTime:      time.Now().Unix(),
	}
	taskManager := tm.GetTaskManager()
	res := false
	for !res {
		res = taskManager.GetBuffer(o.Ot.Id).AddData(tm.OutputData, []*pb.Event{event})
		time.Sleep(1 * time.Second)
	}

}

// a struct of data to key
// tell downstream the data`s key
type Tuple struct {
	Key   []byte
	Value []byte
}

type MapOp interface {
	Map(val []byte) []byte
}

type ReduceOp interface {
	// TODO
	Reduce(va1, val2 []byte) []byte
}

type KeySelector interface {
	GetKey(element []byte) []byte
}

func getSlot(key []byte, parallelism uint) uint {
	hasher := fnv.New32a()
	hasher.Write(key)
	return uint(hasher.Sum32()) % parallelism
}
