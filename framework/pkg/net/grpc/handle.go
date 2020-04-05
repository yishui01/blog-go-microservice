package grpc

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/ecode/transform"
	newmeta "github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	gstatus "google.golang.org/grpc/status"
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
			defer cancel() //note: 服务端在返回时会直接cancel防止内存泄露，那么异步任务传这个ctx的时候需要注意，需要copy出一个不会cancel的ctx才行
		}
		resp, err = handler(ctx, req)
		return resp, transform.FromError(err).Err()
	}
}

// 客户端处理，设置超时时间，转换错误码
func (c *Client) handle() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		var (
			gmd                  = make(metadata.MD)
			conf   *ClientConfig = c.conf
			cancel context.CancelFunc
			p      peer.Peer
		)
		var ec ecode.Codes = ecode.OK
		var timeOpt *TimeoutCallOption

		for _, opt := range opts {
			var tok bool
			timeOpt, tok = opt.(*TimeoutCallOption)
			if tok {
				break
			}
		}
		if timeOpt != nil && timeOpt.Timeout > 0 { //如果客户端设置了这个选项那么强制忽略ctx的剩余超时时间，重新设置Timeout
			ctx, cancel = context.WithTimeout(newmeta.WithContext(ctx), timeOpt.Timeout)
		} else { //否则就设置ctx的超时时间，看剩余时间和conf.Timeout那个短，用短的那个
			_, ctx, cancel = utils.Shrink(ctx, conf.Timeout)
		}

		defer cancel()

		newmeta.Range(ctx,
			func(key string, value interface{}) {
				if valstr, ok := value.(string); ok {
					gmd[key] = []string{valstr}
				}
			},
			newmeta.IsOutgoingKey)

		// merge with old matadata if exists
		if oldmd, ok := metadata.FromOutgoingContext(ctx); ok {
			gmd = metadata.Join(gmd, oldmd)
		}
		ctx = metadata.NewOutgoingContext(ctx, gmd)

		opts = append(opts, grpc.Peer(&p))

		//调用方法，将grpc错误码转换为ecode
		if err = invoker(ctx, method, req, reply, cc, opts...); err != nil {
			gst, _ := gstatus.FromError(err)
			ec = transform.ToEcode(gst)
			err = errors.WithMessage(ec, gst.Message())
		}

		return
	}
}

//server端实际返回的都是grpc的unknow
func GrpcStatusToHttpStatus() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		var (
			p peer.Peer
		)
		opts = append(opts, grpc.Peer(&p))
		var ec ecode.Codes = ecode.OK
		//调用方法，将grpc的status（自定义的永远为unknow）转换为ecode,取出ecode码再转换为http状态码
		if err = invoker(ctx, method, req, reply, cc, opts...); err != nil {
			gst, _ := gstatus.FromError(err)
			ec = transform.ToEcode(gst)
			err = gstatus.New(transform.TogRPCCode(ec), gst.Message()).Err()
		}
		return
	}
}
