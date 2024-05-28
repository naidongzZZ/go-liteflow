package test

import (
	"context"
	"go-liteflow/internal/pkg/operator"
	"go-liteflow/internal/task_manager"
	pb "go-liteflow/pb"
	"plugin"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func InitOperatorPlugin(t *testing.T) {

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
}

func TestInvode(t *testing.T) {

	InitOperatorPlugin(t)

	tm := task_manager.NewTaskManager("", "")
	output := task_manager.NewChannel("output")

	opTaskReduce := &pb.OperatorTask{
		Id: 		"opTaskReduce",
		OpType: 	pb.OpType_Reduce,
		Downstream: []*pb.OperatorTask{{Id: "output"}},
	}
	opTaskReduceInput := task_manager.NewChannel(opTaskReduce.Id)
	
	opTaskMap := &pb.OperatorTask{
		Id: 		"opTaskIdMap1",
		OpType: 	pb.OpType_Map,
		Downstream: []*pb.OperatorTask{{Id: opTaskReduce.Id}},
	}
	opTaskMapInput := task_manager.NewChannel(opTaskMap.Id)

	opTaskMap2 := &pb.OperatorTask{
		Id: 		"opTaskIdMap2",
		OpType: 	pb.OpType_Map,
		Downstream: []*pb.OperatorTask{{Id: opTaskReduce.Id}},
	}
	opTaskMap2Input := task_manager.NewChannel(opTaskMap2.Id)
	tm.RegisterChannel(output, opTaskMapInput, opTaskMap2Input, opTaskReduceInput)


	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm.Invoke(ctx, opTaskMap, opTaskMapInput)
	}()
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm.Invoke(ctx, opTaskMap2, opTaskMap2Input)
	}()
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm.Invoke(ctx, opTaskReduce, opTaskReduceInput)
	}()

	opTaskMapInput.InputCh() <- &pb.Event{
		EventType: pb.EventType_DataOutPut,
		Data: 	  []byte("sss kkk jjj lll www"),
	}
	opTaskMap2Input.InputCh() <- &pb.Event{
		EventType: pb.EventType_DataOutPut,
		Data: []byte("sss kkk jjj"),
	}


	time.Sleep(1 * time.Second)
	opTaskReduceInput.InputCh() <- &pb.Event{
		EventType: pb.EventType_DataSent,
	}

	res := map[string]int {
		"sss": 2,
		"kkk": 2,
		"jjj": 2,
		"lll": 1,
		"www": 1,
	}
	for ev := range output.InputCh() {
		key := string(ev.Key)
		if v, ok := res[key]; ok {
			r , _ := strconv.Atoi(string(ev.Data))
			assert.Equal(t, v, r)
		}
		delete(res, key)
		t.Logf("output: %v \n", ev)
		if len(res) == 0 {
			break
		}
	}
	cancel()
	wg.Wait()
}

func TestInvodeByEventChannel(t *testing.T) {

	InitOperatorPlugin(t)

	tm1 := task_manager.NewTaskManager("127.0.0.1:20020", "")
	tm2 := task_manager.NewTaskManager("127.0.0.1:20021", "")
	
	output := task_manager.NewChannel("output")

	opTaskReduce := &pb.OperatorTask{
		Id: 		"opTaskReduce",
		OpType: 	pb.OpType_Reduce,
		Downstream: []*pb.OperatorTask{{Id: "output"}},
	}
	conn, _ := grpc.Dial("127.0.0.1:20021", grpc.WithInsecure())
	opTaskReduceInput, _ := task_manager.NewChannelWithClient(context.TODO(), tm2.ID(), pb.NewCoreClient(conn))
	
	opTaskMap := &pb.OperatorTask{
		Id: 		"opTaskIdMap1",
		OpType: 	pb.OpType_Map,
		Downstream: []*pb.OperatorTask{{Id: opTaskReduce.Id, TaskManagerId: tm2.ID()}},
	}
	opTaskMapInput := task_manager.NewChannel(opTaskMap.Id)

	tm1.RegisterChannel(opTaskMapInput)
	tm2.RegisterChannel(output, opTaskReduceInput)


	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm2.Start(context.Background())
	}()

	time.Sleep(1 * time.Second)
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm1.Invoke(ctx, opTaskMap, opTaskMapInput)
	}()
	wg.Add(1)
	go func ()  {
		defer wg.Done()
		tm2.Invoke(ctx, opTaskReduce, opTaskReduceInput)
	}()

	opTaskMapInput.InputCh() <- &pb.Event{
		EventType: pb.EventType_DataOutPut,
		Data: 	  []byte("sss sss jjj"),
	}

	time.Sleep(1 * time.Second)
	opTaskReduceInput.InputCh() <- &pb.Event{
		EventType: pb.EventType_DataSent,
	}

	res := map[string]int {
		"sss": 2,
		"jjj": 1,
	}
	for ev := range output.InputCh() {
		key := string(ev.Key)
		if v, ok := res[key]; ok {
			r , _ := strconv.Atoi(string(ev.Data))
			assert.Equal(t, v, r)
		}
		delete(res, key)
		t.Logf("output: %v \n", ev)
		if len(res) == 0 {
			break
		}
	}
	cancel()
	tm2.Stop(context.Background())
	wg.Wait()
}



