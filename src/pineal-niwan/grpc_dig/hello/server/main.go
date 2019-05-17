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
				Usage: "服务端口",
			},
			&cli.BoolFlag{
				Name:  "showGoId",
				Usage: "显示go routine id",
			},
			&cli.BoolFlag{
				Name:  "showCtx",
				Usage: "显示context类型",
			},
			&cli.BoolFlag{
				Name:  "showDeadline",
				Usage: "显示超时相关信息",
			},
			&cli.BoolFlag{
				Name:  "showCtxErr",
				Usage: "显示context错误",
			},
			&cli.IntFlag{
				Name:  "forceErrCode",
				Usage: "强制返回错误",
			},
			&cli.BoolFlag{
				Name:  "forceReturnNil",
				Usage: "强制返回空值",
			},
			&cli.IntFlag{
				Name:  "sleepSecond",
				Usage: "服务停顿秒数,每秒会检查context",
			},
			&cli.IntFlag{
				Name:  "extraSleepSecond",
				Usage: "另一个服务停顿秒数，但不会每秒检查context，在事后检查context",
			},
			&cli.BoolFlag{
				Name:  "needPanic",
				Usage: "服务时制造panic",
			},
			&cli.BoolFlag{
				Name:  "needClose",
				Usage: "服务时关闭Service",
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
		NeedPanic:        c.Bool("needPanic"),
		NeedClose:        c.Bool("needClose"),

		Svr: rpcSvr,
	})
	err = rpcSvr.Serve(ln)
	logrus.Infof("服务器退出时的错误:%+v", err)
	return err
}
