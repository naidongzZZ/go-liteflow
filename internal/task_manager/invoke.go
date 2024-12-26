package task_manager

import (
	"context"
	"errors"
	"go-liteflow/internal/pkg/operator"
	pb "go-liteflow/pb"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

func (tm *taskManager) Invoke(ctx context.Context, opTask *pb.OperatorTask, ch *Channel) (err error) {

	opFn, ok := operator.GetOpFn(opTask.ClientId, opTask.OpId, opTask.OpType)
	if !ok {
		return errors.New("unsupported Operator Func")
	}

	// wait for channel ready
	dsch := make(map[string]*Channel)
	var times, maxTryTimes = 0, 30
	for len(opTask.Downstream) != len(dsch) {
		for _, ds := range opTask.Downstream {
			ch := tm.GetChannel(opTask.ClientId, ds.Id)
			if ch != nil {
				dsch[ds.Id] = ch
			}
		}
		if len(opTask.Downstream) == len(dsch) {
			break
		}
		if times > maxTryTimes {
			slog.Error("downstream channel not ready", slog.Any("dsch", dsch))
			return errors.New("channel not ready")
		}
		time.Sleep(3 * time.Second)
	}

	for {
		select {
		case ev := <-ch.InputCh():
			if ev.EventType == pb.EventType_EtUnknown {
				continue
			}

			slog.Info("operator input.", slog.Any("opTaskId", opTask.Id), slog.Any("event", ev))
			// invoke opFn
			output := opFn(ctx, ev)

			if len(opTask.Downstream) == 0 {
				continue
			}
			slog.Info("operator output.", slog.Any("opTaskId", opTask.Id), slog.Any("events", output))

			for _, oev := range output {
				// TODO distribute target opTaskId
				dsId := opTask.Downstream[0].Id
				ds, ok := dsch[dsId]
				if !ok {
					slog.Error("not found ds channel.", slog.String("optaskId", dsId))
					continue
				}

				ds.InputCh() <- &pb.Event{
					Id:             uuid.NewString(),
					EventType:      pb.EventType_DataOutPut,
					EventTime:      ev.EventTime,
					SourceOpTaskId: opTask.Id,
					TargetOpTaskId: dsId,
					Key:            oev.Key,
					Data:           oev.Data}
			}
		case <-ctx.Done():
			slog.Info("operator done.", slog.Any("opTaskId", opTask.Id), slog.Any("err", ctx.Err()))
			return
		}
	}
}
