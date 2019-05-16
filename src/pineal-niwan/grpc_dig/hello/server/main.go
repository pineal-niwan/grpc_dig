package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"net"
	"os"
	"pineal-niwan/grpc_dig/hello/hello_service"
	"pineal-niwan/grpc_dig/hello/pb"
)

func main() {
	app := &cli.App{
		Name:    "gRPC example",
		Version: "1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "port",
				Usage: "listen port",
			},
			&cli.BoolFlag{
				Name: "showGoId",
			},
			&cli.BoolFlag{
				Name: "showCtx",
			},
			&cli.BoolFlag{
				Name: "showDeadline",
			},
			&cli.BoolFlag{
				Name: "showCtxErr",
			},
			&cli.IntFlag{
				Name: "forceErrCode",
			},
			&cli.BoolFlag{
				Name: "forceReturnNil",
			},
			&cli.IntFlag{
				Name: "sleepSecond",
			},
			&cli.IntFlag{
				Name: "extraSleepSecond",
			},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Errorf("run error :%+v\n", err)
	}
}

func run(c *cli.Context) error {
	ln, err := net.Listen("tcp", c.String("port"))
	if err != nil {
		return err
	}

	rpcSvr := grpc.NewServer()
	pb.RegisterHelloServiceServer(rpcSvr, &hello_service.HelloService{
		ShowGoId:         c.Bool("showGoId"),
		ShowCtx:          c.Bool("showCtx"),
		ShowDeadLine:     c.Bool("showDeadline"),
		ShowCtxErr:       c.Bool("showCtxErr"),
		ForceErrCode:     c.Int("forceErrCode"),
		ForceReturnNil:   c.Bool("forceReturnNil"),
		SleepSecond:      c.Int("sleepSecond"),
		ExtraSleepSecond: c.Int("extraSleepSecond"),
	})
	err = rpcSvr.Serve(ln)
	return err
}
