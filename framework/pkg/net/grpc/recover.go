package grpc

import (
	"context"
	"fmt"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
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
				log.SugarWithContext(ctx).Errorf("grpc server panic: %v\n%v\n%s\n", req, rerr, buf)
				err = status.Errorf(codes.Internal, "服务器内部错误")
			}
		}()
		return handler(ctx, req)
	}

}

// recovery return a client interceptor  that recovers from any panics.
func (c *Client) recovery() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				rs := runtime.Stack(buf, false)
				if rs > size {
					rs = size
				}
				buf = buf[:rs]
				pl := fmt.Sprintf("grpc client panic: %v\n%v\n%v\n%s\n", req, reply, rerr, buf)
				fmt.Fprintf(os.Stderr, pl)
				log.ZapWithContext(ctx).Error(pl)
				err = ecode.ServerErr
			}
		}()
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}
