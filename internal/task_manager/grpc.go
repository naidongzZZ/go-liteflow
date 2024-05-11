package task_manager


import (
	"go-liteflow/internal/core"
)

type grpcServer struct {
	core.Comm

	tm *taskManager
}
