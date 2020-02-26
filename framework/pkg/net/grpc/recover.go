package grpc

import (
	"context"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
)

func (s *GrpcServer) reovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				rs := runtime.Stack(buf, false)
				if rs > size {
					rs = size
				}
				buf = buf[:rs]
				log.SugarLogger.Infof("grpc server panic: %v\n%v\n%s\n", req, rerr, buf)
				err = status.Errorf(codes.Internal, "服务器内部错误")
			}
		}()
		return handler(ctx,req)
	}

}
