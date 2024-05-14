package main

import (
	"context"
	"errors"
	"go-liteflow/internal/coordinator"
	"go-liteflow/internal/task_manager"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "run",
			Usage: "task_manager or coordinator",
			Value: "task_manager",
		},
		cli.StringFlag{
			Name:  "addr",
			Usage: "App Address",
			Value: ":20020",
		},
		cli.StringFlag{
			Name:  "coord_addr",
			Usage: "Coordinator Address",
		},
	}

	app.Action = func(ctx *cli.Context) error {

		runMode := ctx.String("run")
		addr := ctx.String("addr")
		if len(addr) == 0 {
			return errors.New("addr is nil")
		}
		
		coordAddr := ctx.String("coord")
		log.Printf("%s start on %s ...", runMode, addr)

		if runMode == "task_manager" {
			tm := task_manager.NewTaskManager(addr, coordAddr)
			tm.Start(context.Background())
		} else if runMode == "coordinator" {
			co := coordinator.NewCoordinator(addr)
			co.Start(context.Background())
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

}
