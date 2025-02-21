package task_manager

import (
	"context"
	"log/slog"
	"runtime/debug"
	"sync"
)

var (
	RunnerPool = newRunnerPool()
)

type runnerPool struct {
	wg   sync.WaitGroup
}

func newRunnerPool() *runnerPool {
	return &runnerPool{
	}
}

func (rp *runnerPool) Run(ctx context.Context, fn func(c context.Context)) {
	rp.wg.Add(1)
	go func() {
		defer rp.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in runner", slog.Any("err", r), slog.Any("stack", string(debug.Stack())))
			}
		}()
		fn(ctx)
	}()
}

func (rp *runnerPool) Wait() {
	rp.wg.Wait()
}

