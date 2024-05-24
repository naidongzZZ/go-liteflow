package test

import (
	"context"
	"go-liteflow/internal/pkg/operator"
	"go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"plugin"
	"sync"
	"testing"
	"time"
)

func TestInvode(t *testing.T) {

	p, err := plugin.Open("plugins/mr.so")
	if err != nil {
		t.Fatal(err)
	}

	sym, err := p.Lookup("MapOpFn")
	if err != nil {
		t.Fatal(err)
	}
	MapOpFn := sym.(func(opId string) operator.OpFn)

	sym2, err := p.Lookup("ReduceOpFn")
	if err != nil {
		t.Fatal(err)
	}
	ReduceOpFn := sym2.(func(opId string) operator.OpFn)

	operator.RegisterOpFn("", pb.OpType_Map, MapOpFn)
	operator.RegisterOpFn("", pb.OpType_Reduce, ReduceOpFn)

	tm := task_manager.NewTaskManager("", "")
	in := make(chan *pb.Event, 1000)
	tmp := make(chan *pb.Event, 1000)
	out := make(chan *pb.Event, 1000)

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



