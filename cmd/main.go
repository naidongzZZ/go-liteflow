package main

import (
	"context"
	"go-liteflow/internal/coordinator"
	"go-liteflow/internal/task_manager"
	"os"

	"log/slog"

	"github.com/urfave/cli"
)

const (
	RunModeTaskManager string = "tm"
	RunModeCoordinator string = "co"
)

// usage: go run cmd/main.go help
func main() {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	app := cli.NewApp()
	app.Name = "liteflow"
	app.Usage = "🐣 lightweight distributed stream processing"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "mode",
			Usage: "Start (tm) task_manager or (co) coordinator",
		},
		cli.StringFlag{
			Name:  "addr",
			Usage: "App's Address",
		},
		cli.StringFlag{
			Name:  "co_addr",
			Usage: "Coordinator's Address",
		},
	}

	app.Action = func(ctx *cli.Context) error {

		runMode, addr := ctx.String("mode"), ctx.String("addr")
		if len(addr) == 0 {
			slog.Error("addr is illegal", slog.String("addr", addr))
			return nil
		}

		slog.Info("app start.",slog.String("mode", runMode),slog.String("addr", addr))

		if runMode == RunModeTaskManager {
			coordAddr := ctx.String("co_addr")

			tm := task_manager.NewTaskManager(addr, coordAddr)
			slog.Info("task_manager info.", slog.String("ID", tm.ID()))

			tm.Start(context.Background())
		} else if runMode == RunModeCoordinator {
			co := coordinator.NewCoordinator(addr)
			slog.Info("coordinator info.", slog.String("ID", co.ID()))

			co.Start(context.Background())
			slog.Info("coordinator info, coordinator exist")
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("server exit", slog.Any("err", err))
	}

}
