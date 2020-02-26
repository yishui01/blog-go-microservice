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
			logFields = append(logFields, zap.String("args", req.(fmt.Stringer).String()))
		}
		if err != nil {
			logFields = append(logFields, zap.String("error", err.Error()), zap.String("stack", fmt.Sprintf("%+v", err)))
		}
		logFn(int(code), duration)("grpc-custom-log:", logFields...)
		return resp, err
	}
}
