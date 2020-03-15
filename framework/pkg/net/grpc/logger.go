package grpc

import (
	"context"
	"fmt"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"strconv"
	"time"
)

// Log Flag
const (
	// disable all log.
	LogFlagDisable = 1 << iota
	// disable print args on log.
	LogFlagDisableArgs
	// disable info level log.
	LogFlagDisableInfo
)

func logFn(code int, dt time.Duration) func(msg string, f ...zap.Field) {
	switch code {
	case int(codes.OK):
		return log.ZapLogger.Info
	case int(codes.Internal):
		return log.ZapLogger.Error
	default:
		return log.ZapLogger.Warn
	}
}

func serverLog(logFlag int8) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
		_metricServerReqDur.WithLabelValues(info.FullMethod, caller).Observe(float64(duration / time.Millisecond))
		_metricServerReqCodeTotal.WithLabelValues(info.FullMethod, caller, strconv.Itoa(int(code))).Inc()

		if logFlag&LogFlagDisable != 0 {
			return resp, err
		}

		// TODO: find better way to deal with slow log.
		if logFlag&LogFlagDisableInfo != 0 && err == nil && duration < 500*time.Millisecond {
			return resp, err
		}

		logFields := []zap.Field{
			zap.String("user", caller),
			zap.String("ip", remoteIP),
			zap.String("path", info.FullMethod),
			zap.Int("ret", int(code)),
			zap.Float64("ts", duration.Seconds()),
			zap.Float64("timeout_quota", quota),
			zap.String("source", "grpc-access-log"),
		}

		if logFlag&LogFlagDisableArgs == 0 {
			// TODO: it will panic if someone remove String method from protobuf message struct that auto generate from protoc.
			//logFields = append(logFields, zap.String("args", req.(fmt.Stringer).String())) //打不出中文，V2才支持
			logFields = append(logFields, zap.String("args", fmt.Sprintf("%#+v", req))) //只能反射了

		}
		if err != nil {
			logFields = append(logFields, zap.String("error", err.Error()), zap.String("stack", fmt.Sprintf("%+v", err)))
		}
		logFn(int(code), duration)("grpc-server-log:", logFields...)
		return resp, err
	}
}

//grpc Client的logFlag只能从grpc的DialOption或者CallOption中提取，直接传没法传，所以只能封装下
type logOption struct {
	grpc.EmptyDialOption
	grpc.EmptyCallOption
	flag int8
}

// WithLogFlag disable client access log.
func WithLogFlag(flag int8) grpc.CallOption {
	return logOption{flag: flag}
}

// WithDialLogFlag set client level log behaviour.
func WithDialLogFlag(flag int8) grpc.DialOption {
	return logOption{flag: flag}
}

func extractLogCallOption(opts []grpc.CallOption) (flag int8) {
	for _, opt := range opts {
		if logOpt, ok := opt.(logOption); ok {
			return logOpt.flag
		}
	}
	return
}

func extractLogDialOption(opts []grpc.DialOption) (flag int8) {
	for _, opt := range opts {
		if logOpt, ok := opt.(logOption); ok {
			return logOpt.flag
		}
	}
	return
}

// clientLogging warden grpc logging
func clientLogging(dialOptions ...grpc.DialOption) grpc.UnaryClientInterceptor {
	defaultFlag := extractLogDialOption(dialOptions)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		logFlag := extractLogCallOption(opts) | defaultFlag

		startTime := time.Now()
		var peerInfo peer.Peer
		opts = append(opts, grpc.Peer(&peerInfo))

		// invoker requests
		err := invoker(ctx, method, req, reply, cc, opts...)

		// after request
		code := ecode.Cause(err).Code()
		duration := time.Since(startTime)

		// monitor
		_metricClientReqDur.WithLabelValues(method).Observe(float64(duration / time.Millisecond))
		_metricClientReqCodeTotal.WithLabelValues(method, strconv.Itoa(code)).Inc()

		if logFlag&LogFlagDisable != 0 {
			return err
		}
		// TODO: find better way to deal with slow log.
		if logFlag&LogFlagDisableInfo != 0 && err == nil && duration < 500*time.Millisecond {
			return err
		}

		logFields := []zap.Field{
			zap.String("path", method),
			zap.Int("ret", code),
			zap.Float64("ts", duration.Seconds()),
			zap.String("source", "grpc-access-log"),
		}

		if peerInfo.Addr != nil {
			logFields = append(logFields, zap.String("ip", peerInfo.Addr.String()))
		}
		if logFlag&LogFlagDisableArgs == 0 {
			// TODO: it will panic if someone remove String method from protobuf message struct that auto generate from protoc.
			//logFields = append(logFields, zap.String("args", req.(fmt.Stringer).String()))
			logFields = append(logFields, zap.String("args", fmt.Sprintf("%#+v", req)))
		}
		if err != nil {
			logFields = append(logFields, zap.String("error", err.Error()), zap.String("stack", fmt.Sprintf("%+v", err)))
		}

		logFn(code, duration)("grpc-client-log:", logFields...)
		return err
	}
}
