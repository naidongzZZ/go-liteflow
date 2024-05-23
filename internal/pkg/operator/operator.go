package operator

import (
	"context"
	"fmt"
	pb "go-liteflow/pb"
	"sync"
)

var (
	_mux        sync.Mutex
	_operatorFn = make(map[string]OpFnGenerator)
)

type (
	OpFn func(context.Context, ...*pb.Event) []*pb.Event

	OpFnGenerator func(opId string) OpFn
)

func OperatorFnKey(opId string, opType pb.OpType) string {
	return fmt.Sprintf("%s_%s", opType, opId)
}

func RegisterOpFn(opId string, opType pb.OpType, fn OpFnGenerator) {
	_mux.Lock()
	defer _mux.Unlock()
	_operatorFn[OperatorFnKey(opId, opType)] = fn
}

func UnregisterOpFn(opId string, opType pb.OpType) {
	_mux.Lock()
	defer _mux.Unlock()
	delete(_operatorFn, OperatorFnKey(opId, opType))
}

func GetOpFn(cId, opId string, opType pb.OpType) (f OpFn, ok bool) {

	_mux.Lock()
	defer _mux.Unlock()

	fnGenerator, ok := _operatorFn[OperatorFnKey(opId, opType)]
	if ok {
		return fnGenerator(opId), ok
	}
	return
}
