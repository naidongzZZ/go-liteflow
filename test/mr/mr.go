package main

import (
	"go-liteflow/internal/pkg/log"
	"os"

	"github.com/urfave/cli"
)


func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "op",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		op := ctx.String("op")

		log.Infof("run op: %s", op)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Errorf("server exit: %v", err)
	}

}
