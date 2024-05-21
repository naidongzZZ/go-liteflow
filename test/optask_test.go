package test

import (
	"context"
	"go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvode(t *testing.T) {

	tm := task_manager.NewTaskManager("", "")
	in := make(chan *pb.Event, 10)
	tmp := make(chan *pb.Event, 10)
	out := make(chan *pb.Event, 10)

	opTaskMap := &pb.OperatorTask{
		Id: 		"opTaskIdMap1",
		OpId: 		"opIdMap1",
		OpType: 	pb.OpType_Map,
	}
	opTaskMap2 := &pb.OperatorTask{
		Id: 		"opTaskIdMap2",
		OpId: 		"opIdMap2",
		OpType: 	pb.OpType_Map,
	}
	opTaskReduce := &pb.OperatorTask{
		Id: 		"opTaskReduce",
		OpId: 		"opIdReduce1",
		OpType: 	pb.OpType_Reduce,
		Downstream: []*pb.OperatorTask{{Id: "output"}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		err := tm.Invoke(ctx, opTaskMap, in, tmp)
		t.Logf("invoke err: %v, Id: %v", err, opTaskMap.Id)
	}()
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		err := tm.Invoke(ctx, opTaskMap2, in, tmp)
		t.Logf("invoke err: %v, Id: %v", err, opTaskMap2.Id)
	}()
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		err := tm.Invoke(ctx, opTaskReduce, tmp, out)
		t.Logf("invoke err: %v, Id: %v", err, opTaskReduce.Id)
	}()

	in <- &pb.Event{
		EventType: pb.EventType_DataOutPut,
		Data: 	  []byte("data"),
	}
	in <- &pb.Event{
		EventType: pb.EventType_DataOutPut,
		Data: []byte("data2"),
	}
	
	wg.Wait()
	assert.Equal(t, 2, len(out))
}