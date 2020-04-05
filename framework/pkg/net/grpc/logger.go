package grpc

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	uuid "github.com/satori/go.uuid"
	"github.com/uber/jaeger-client-go"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"time"
)

func logFn(ctx context.Context, code int, dt time.Duration) func(msg string, f ...zap.Field) {
	switch code {
	case int(codes.OK):
		return log.ZapWithContext(ctx).Info
	case int(codes.Internal):
		return log.ZapWithContext(ctx).Error
	default:
		return log.ZapWithContext(ctx).Warn
	}
}

func serverLog() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//把网关注入的traceId提取出来，作为log的requestId注回context中
		var reqId string
		span := opentracing.SpanFromContext(ctx)

		if sc, ok := span.Context().(jaeger.SpanContext); ok {
			reqId = sc.TraceID().String()
		} else {
			reqId = "uuid-" + uuid.NewV4().String()
		}
		ctx = log.NewContext(ctx, zap.String("requestId", reqId))

		startTime := time.Now()
		caller := metadata.String(ctx, metadata.Caller)
		if caller == "" {
			caller = "no_user"
		}
		var remoteIP string
		if peerInfo, ok := peer.FromContext(ctx); ok {
			remoteIP = peerInfo.Addr.String()
		}
		var quota float64
		if deadline, ok := ctx.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}
		// call server handler
		resp, err = handler(ctx, req)

		//查看错误码是否为grpc的错误码
		code := ecode.Cause(err).Code()

		duration := time.Since(startTime)
		// monitor
		//_metricServerReqDur.WithLabelValues(info.FullMethod, caller).Observe(float64(duration / time.Millisecond))
		//_metricServerReqCodeTotal.WithLabelValues(info.FullMethod, caller, strconv.Itoa(int(code))).Inc()

		logFields := []zap.Field{
			zap.String("user", caller),
			zap.String("ip", remoteIP),
			zap.String("path", info.FullMethod),
			zap.Int("ret", code),
			zap.Float64("ts", duration.Seconds()),
			zap.Float64("timeout_quota", quota),
			zap.String("source", "grpc-access-log"),
		}

		//logFields = append(logFields, zap.String("args", req.(fmt.Stringer).String())) //打不出中文，V2才支持
		logFields = append(logFields, zap.String("args", fmt.Sprintf("%#+v", req))) //只能反射了

		if err != nil {
			logFields = append(logFields, zap.String("error", err.Error()), zap.String("stack", fmt.Sprintf("%+v", err)))
		}
		logFn(ctx, code, duration)("grpc-server-log:", logFields...)
		return resp, err
	}
}

// clientLogging warden grpc logging
func clientLogging() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		var peerInfo peer.Peer
		opts = append(opts, grpc.Peer(&peerInfo))

		// invoker requests
		err := invoker(ctx, method, req, reply, cc, opts...)

		// after request
		code := ecode.Cause(err).Code()
		duration := time.Since(startTime)

		// monitor
		//_metricClientReqDur.WithLabelValues(method).Observe(float64(duration / time.Millisecond))
		//_metricClientReqCodeTotal.WithLabelValues(method, strconv.Itoa(code)).Inc()

		logFields := []zap.Field{
			zap.String("path", method),
			zap.Int("ret", code),
			zap.Float64("ts", duration.Seconds()),
			zap.String("source", "grpc-access-log"),
		}

		if peerInfo.Addr != nil {
			logFields = append(logFields, zap.String("ip", peerInfo.Addr.String()))
		}
		logFields = append(logFields, zap.String("args", fmt.Sprintf("%#+v", req)))

		if err != nil {
			logFields = append(logFields, zap.String("error", err.Error()), zap.String("stack", fmt.Sprintf("%+v", err)))
		}

		logFn(ctx, code, duration)("grpc-client-log:", logFields...)
		return err
	}
}
