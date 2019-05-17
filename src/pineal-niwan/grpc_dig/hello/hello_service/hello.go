package hello_service

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"pineal-niwan/grpc_dig/go_routine_id"
	"pineal-niwan/grpc_dig/hello/pb"
	"reflect"
	"time"
)

type HelloService struct {
	//是否显示go id
	ShowGoId bool
	//是否显示ctx类型
	ShowCtx bool
	//是否显示超时
	ShowDeadLine bool
	//是否显示context error
	ShowCtxErr bool
	//是否强制返回错误码
	ForceErrCode int
	//是否强制返回空
	ForceReturnNil bool
	//指定的sleep秒数
	SleepSecond int
	//额外指定的sleep秒数
	ExtraSleepSecond int
	//服务时是否制造panic
	NeedPanic bool
	//服务时是否关闭Service
	NeedClose bool

	Svr *grpc.Server
}

type ErrSelf struct {
	Code int
}

func (err ErrSelf) Error() string {
	return fmt.Sprintf("code:%+v", err.Code)
}

func (h *HelloService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	var rsp *pb.HelloResponse

	if h.ForceReturnNil {
		rsp = nil
	} else {
		rsp = &pb.HelloResponse{
			Reply: req.Greeting + "rsp",
		}
	}

	if h.ShowGoId {
		logrus.Infof("go routine %d\n", go_routine_id.Goid())
	}
	if h.ShowCtx {
		logrus.Infof("reflect of ctx %+v\n", reflect.TypeOf(ctx))
	}
	if h.ShowDeadLine {
		deadline, ok := ctx.Deadline()
		if ok {
			now := time.Now()
			logrus.Infof("current time:%+v  deadline:%+v   delta:%+v\n", now, deadline, deadline.Sub(now))
		} else {
			logrus.Infof("there is no deadline set here")
		}
	}

	done := ctx.Done()

	if h.ExtraSleepSecond > 0 {
		time.Sleep(time.Second * time.Duration(h.ExtraSleepSecond))
		select {
		case <-done:
			err := ctx.Err()
			if h.ShowCtxErr {
				logrus.Infof("额外超时后context err is %+v\n", err)
			}
			logrus.Infof("额外超时后done后的返回 %+v, -- %+v", rsp, err)
			return rsp, err
		default:
		}
	}

	if h.SleepSecond > 0 {
		for i := 0; i < h.SleepSecond; i++ {
			//每秒钟检查是否有cancel
			time.Sleep(time.Second)
			select {
			case <-done:
				err := ctx.Err()
				if h.ShowCtxErr {
					logrus.Infof("context err is %+v\n", err)
				}
				logrus.Infof("done后的返回 %+v, -- %+v", rsp, err)
				return rsp, err
			default:
			}
		}
	}

	var err error

	if h.ForceErrCode > 0 {
		err = &ErrSelf{Code: h.ForceErrCode}
	}
	logrus.Infof("normal返回 %+v, -- %+v", rsp, err)

	if h.NeedPanic {
		panic("panic测试")
	}

	if h.NeedClose {
		go h.Svr.GracefulStop()
		//h.Svr.GracefulStop()
	}

	return rsp, err
}
