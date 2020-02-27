package grpc

import (
	"context"
	"github.com/zuiqiangqishao/framework/pkg/ecode/transform"
	"google.golang.org/grpc"
	"time"
)

//处理当前grpc配置的超时时间和上游传下来的ctx剩余时间，设置当前请求的最终超时时间
func (s *GrpcServer) handle() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			cancel func()
		)
		s.mutex.RLock()
		conf := s.conf
		s.mutex.RUnlock()
		//将当前grpc剩余的超时时间和配置的超时时间比较，取较小的那个
		timeout := conf.Timeout
		if de, ok := ctx.Deadline(); ok {
			ctimeout := time.Until(de)
			if ctimeout-time.Millisecond*20 > 0 {
				ctimeout = ctimeout - time.Millisecond*20
			}
			if timeout > ctimeout {
				timeout = ctimeout
			}
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		resp, err = handler(ctx, req)
		return resp, transform.FromError(err).Err()
	}
}
