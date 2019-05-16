package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"os"
	"pineal-niwan/grpc_dig/hello/pb"
	"time"
)

func main() {
	app := &cli.App{
		Name:    "gRPC example",
		Version: "1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url",
				Usage: "rpc url",
			},
			&cli.IntFlag{
				Name: "count",
			},
			&cli.BoolFlag{
				Name: "async",
			},
			&cli.Int64Flag{
				Name: "timeout",
			},
			&cli.Int64Flag{
				Name: "intervalSecond",
			},
			&cli.BoolFlag{
				Name: "retry",
			},
			&cli.StringFlag{
				Name: "deadline",
			},
			&cli.BoolFlag{
				Name: "cancelInHalfTime",
			},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Errorf("run error :%+v\n", err)
	}
}

type RPCClient struct {
	*grpc.ClientConn
	pb.HelloServiceClient
}

func run(c *cli.Context) error {
	var deadline time.Time

	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})

	ctxDial, cancelDial := context.WithTimeout(context.Background(), time.Second*2)
	conn, err := grpc.DialContext(ctxDial, c.String("url"), grpc.WithInsecure())
	cancelDial()

	if err != nil {
		logrus.Infof("连接RPC server错误:%+v\n", err)
		return err
	}

	client := pb.NewHelloServiceClient(conn)
	rpcClient := &RPCClient{
		ClientConn:         conn,
		HelloServiceClient: client,
	}

	count := c.Int("count")
	timeout := time.Duration(c.Int64("timeout")) * time.Second
	deadlineStr := c.String("deadline")
	cancelInHalfTime := c.Bool("cancelInHalfTime")

	if deadlineStr != "" {
		deadline, err = time.ParseInLocation("2006-01-02 15:04:05", deadlineStr, time.Local)
		if err != nil {
			return err
		}
	}

	if c.Bool("async") {
		asyncCall(rpcClient, count, timeout, deadline, cancelInHalfTime)
	} else {
		err = syncCall(rpcClient, count, timeout, deadline, cancelInHalfTime,
			c.Int64("intervalSecond"), c.Bool("retry"))
	}
	return err
}

func callRPC(client *RPCClient, i int, timeout time.Duration, deadline time.Time, cancelInHalfTime bool) error {
	var rsp *pb.HelloResponse
	var err error

	req := &pb.HelloRequest{
		Greeting: fmt.Sprintf("test%d", i),
	}

	if timeout > 0 {
		ctxCaller, cancleCaller := context.WithTimeout(context.Background(), timeout)
		if cancelInHalfTime {
			go func() {
				time.Sleep(timeout / 2)
				cancleCaller()
			}()
		}
		rsp, err = client.SayHello(ctxCaller, req)
		cancleCaller()
	} else if deadline.IsZero() {
		rsp, err = client.SayHello(context.Background(), req)
	} else {
		now := time.Now()
		logrus.Infof("client now:%+v deadline:%+v delta:%+v\n", now, deadline, deadline.Sub(now))
		ctxCaller, cancleCaller := context.WithDeadline(context.Background(), deadline)
		rsp, err = client.SayHello(ctxCaller, req)
		cancleCaller()
	}

	logrus.Infof("RPC返回 %+v -- %+v\n", rsp, err)

	if rsp == nil {
		logrus.Infof("rsp is nil \n")
	} else {
		logrus.Infof("rsp is %+v\n", *rsp)
	}

	if err != nil {
		logrus.Infof("call error %+v\n", err)
		gRpcErr, ok := status.FromError(err)
		if ok {
			logrus.Infof("gRpcErr code:%+v err:%+v detail:%+v conn state:%+v target:%+v\n",
				gRpcErr.Code(), gRpcErr.Err(), gRpcErr.Details(), client.GetState(), client.Target())
		}
	}

	return err
}

func syncCall(client *RPCClient, count int, timeout time.Duration, deadline time.Time,
	cancelInHalfTime bool, intervalSecond int64, retry bool) error {
	for i := 0; i < count; i++ {
		err := callRPC(client, i, timeout, deadline, cancelInHalfTime)
		if intervalSecond > 0 {
			//间隔多少秒后继续
			time.Sleep(time.Duration(intervalSecond) * time.Second)

			if err != nil {
				logrus.Infof("再次查看状态 %+v\n", err)
				gRpcErr, ok := status.FromError(err)
				if ok {
					logrus.Infof("gRpcErr code:%+v err:%+v detail:%+v conn state:%+v target:%+v\n",
						gRpcErr.Code(), gRpcErr.Err(), gRpcErr.Details(), client.GetState(), client.Target())
				}
			}
		} else if !retry {
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func asyncCall(client *RPCClient, count int, timeout time.Duration, deadline time.Time, cancelInHalfTime bool) {
	done := make(chan struct{}, count)

	caller := func(i int) {
		err := callRPC(client, i, timeout, deadline, cancelInHalfTime)
		if err != nil {
			logrus.Infof("async call err:%+v\n", err)
		}
		done <- struct{}{}
	}

	for i := 0; i < count; i++ {
		go caller(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}
}
