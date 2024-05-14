package main

import (
	"context"
	"go-liteflow/internal/task_manager"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Usage: "App Address",
			Value: ":20020",
		},
		cli.StringFlag{
			Name:  "coord",
			Usage: "Coordinator Address",
		},
	}

	app.Action = func(ctx *cli.Context) error {

		addr := ctx.String("addr")
		coordAddr := ctx.String("coord")
		log.Println("server start...")

		tm := task_manager.NewTaskManager(addr, coordAddr)

		tm.Start(context.Background())

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

}
