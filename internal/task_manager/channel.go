package task_manager

import (
	"context"
	pb "go-liteflow/pb"
)

type Channel struct {
	opTaskId string
	inputCh  chan *pb.Event
	outputCh chan *pb.Event
}

// todo channel options
func NewChannel(opTaskId string) *Channel {
	c := &Channel{
		opTaskId: opTaskId,
		inputCh:  make(chan *pb.Event, 1000),
		outputCh: make(chan *pb.Event, 1000),
	}
	return c
}

func (ch *Channel) InputCh() chan *pb.Event {
	return ch.inputCh
}

func (ch *Channel) OutputCh() chan *pb.Event {
	return ch.outputCh
} 

func (ch *Channel) PutInputCh(ctx context.Context, ev *pb.Event) (ok bool) {
	ch.inputCh <- ev
	return true
}

func (ch *Channel) GetInputCh(ctx context.Context) (ev *pb.Event, ok bool) {
	ev = <-ch.inputCh
	return ev, true
}

func (ch *Channel) PutOutputCh(ctx context.Context, ev *pb.Event) (ok bool) {
	ch.outputCh <- ev
	return true
}

func (ch *Channel) GetOutputCh(ctx context.Context) (ev *pb.Event, ok bool) {
	ev = <-ch.outputCh
	return
}
