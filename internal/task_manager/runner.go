package task_manager

import (
	"context"
	"go-liteflow/internal/pkg/log"
	"runtime/debug"
	"sync"
)

var (
	rootwg              = sync.WaitGroup{}
	rootCtx, rootCancel = context.WithCancel(context.Background())
)

type Runnable interface {
}

type LaunchFn func(context.Context) (error)

func GracefulRun(f LaunchFn) {
	rootwg.Add(1)
	go func() {
		defer rootwg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in runner. err: %v, stack: %s", r, string(debug.Stack()))
			}
		}()

		err := f(rootCtx)
		if err != nil {
			log.Warnf("runner exit. err: %v", err)
			return
		}
	}()
}

func GracefulStop() {
	log.Infof("graceful stoping")
	rootCancel()
	log.Infof("wait for cancel")
	rootwg.Wait()
}
