package task_manager

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestGracefulRun(t *testing.T) {
	GracefulRun(
		func(c context.Context) (err error) {
			cmd := exec.CommandContext(c, "ping", "127.0.0.1")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Cancel = func() error {
				return cmd.Process.Signal(syscall.SIGTERM)
			}
			return cmd.Run()
		},
	)

	time.Sleep(3 * time.Second)
	GracefulStop()
}
