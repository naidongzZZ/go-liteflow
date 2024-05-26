package task_manager

import (
	"context"
	pb "go-liteflow/pb"
	"log/slog"
)

type Channel struct {
	taskManagerId string
	opTaskId      string
	inputCh       chan *pb.Event
	client        pb.CoreClient
}

// TODO channel options
func NewChannel(opTaskId string) *Channel {
	c := &Channel{
		opTaskId: opTaskId,
		inputCh:  make(chan *pb.Event, 1000),
	}
	return c
}

func NewChannelWithClient(
	ctx context.Context,
	taskManagerId string,
	client pb.CoreClient) (c *Channel, err error) {
	c = &Channel{
		taskManagerId: taskManagerId,
		inputCh:       make(chan *pb.Event, 5000),
		client:        client,
	}

	go func() {
		stream, e := c.client.DirectedEventChannel(ctx)
		if e != nil {
			err = e
			slog.Error("create event channel.", slog.Any("err", e))
			return
		}

		for {
			select {
			case ev := <-c.inputCh:
				req := &pb.EventChannelReq{}
				req.Events = append(req.Events, ev)

				if e := stream.Send(req); e != nil {
					slog.Error("send to event channel.", slog.Any("err", e))
				}
			case <-ctx.Done():
				slog.Info("recv done.", slog.Any("err", ctx.Err()))
				return
			}
		}
	}()
	return c, nil
}

func (ch *Channel) InputCh() chan *pb.Event {
	return ch.inputCh
}

func (ch *Channel) PutInputCh(ctx context.Context, ev *pb.Event) (ok bool) {
	ch.inputCh <- ev
	return true
}
