package test

import (
	"context"
	"encoding/json"
	"go-liteflow/internal/pkg/operator"
	"go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInvode(t *testing.T) {

	operator.RegisterOpFn("", pb.OpType_Map, MapOpFn)
	operator.RegisterOpFn("", pb.OpType_Reduce, ReduceOpFn)

	tm := task_manager.NewTaskManager("", "")
	in := make(chan *pb.Event, 10)
	tmp := make(chan *pb.Event, 10)
	out := make(chan *pb.Event, 10)

	opTaskMap := &pb.OperatorTask{
		Id: 		"opTaskIdMap1",
		OpType: 	pb.OpType_Map,
		Downstream: []*pb.OperatorTask{{Id: "output"}},
	}
	opTaskMap2 := &pb.OperatorTask{
		Id: 		"opTaskIdMap2",
		OpType: 	pb.OpType_Map,
		Downstream: []*pb.OperatorTask{{Id: "output"}},
	}
	opTaskReduce := &pb.OperatorTask{
		Id: 		"opTaskReduce",
		OpType: 	pb.OpType_Reduce,
		Downstream: []*pb.OperatorTask{{Id: "output"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm.Invoke(ctx, opTaskMap, in, tmp)
	}()
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm.Invoke(ctx, opTaskMap2, in, tmp)
	}()
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm.Invoke(ctx, opTaskReduce, tmp, out)
	}()

	in <- &pb.Event{
		EventType: pb.EventType_DataOutPut,
		Data: 	  []byte("sss kkk jjj lll www"),
	}
	in <- &pb.Event{
		EventType: pb.EventType_DataOutPut,
		Data: []byte("sss kkk jjj"),
	}

	wg.Wait()

	
	//assert.Equal(t, 2, len(out))
}

func MapOpFn() operator.OpFn {
	return func(ctx context.Context, input []byte) (output []byte) {
	
		res := make([]map[string]int, 0)
	
		for _, word := range strings.Split(string(input) , " ") {
			res = append(res, map[string]int {strings.Trim(word, " "): 1})
		}
	
		bytes, _ := json.Marshal(res)
		return bytes
	}
}



func ReduceOpFn() operator.OpFn {
	var m = make(map[string]int)

	return func(ctx context.Context, input []byte) (output []byte){
		in := make([]map[string]int, 0)
		_ = json.Unmarshal(input, &in)
		for _, kvs := range in {
			for k, v := range kvs {
				m[k] +=v
			}
		}
		output, _ = json.Marshal(m)
		return output
	}
}