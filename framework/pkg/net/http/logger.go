package khttp

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	uuid "github.com/satori/go.uuid"
	"github.com/uber/jaeger-client-go"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// Logger is logger  middleware
func Logger() HandlerFunc {
	const noUser = "no_user"
	return func(c *Context) {
		now := time.Now()
		req := c.Request
		ip := metadata.String(c, metadata.RemoteIP)
		path := req.URL.Path
		params := req.Form

		//把网关注入的traceId提取出来，作为log的requestId注回context中
		var reqId string
		span := opentracing.SpanFromContext(c.Context) //trace一律从Context中获取，req的header头中没有trace
		if sc, ok := span.Context().(jaeger.SpanContext); ok {
			reqId = sc.TraceID().String()
		} else {
			reqId = "uuid-" + uuid.NewV4().String()
		}
		c.Context = log.NewContext(c.Context, zap.String("requestId", reqId)) //注入log

		var quota float64
		if deadline, ok := c.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}
		c.Next()

		err := c.Error
		cerr := ecode.Cause(err)
		dt := time.Since(now)

		caller := metadata.String(c, metadata.Caller)
		if caller == "" {
			caller = noUser
		}

		if len(c.RoutePath) > 0 {
			_metricServerReqCodeTotal.WithLabelValues(c.RoutePath[1:], caller, req.Method, strconv.FormatInt(int64(cerr.Code()), 10)).Inc()
			_metricServerReqDur.WithLabelValues(c.RoutePath[1:], caller, req.Method).Observe(float64(int64(dt / time.Millisecond)))
		}

		lf := log.ZapWithContext(c).Info
		errmsg := ""
		isSlow := dt >= (time.Millisecond * 500)
		if err != nil {
			errmsg = err.Error()
			lf = log.ZapWithContext(c).Error
			if cerr.Code() > 0 {
				lf = log.ZapWithContext(c).Warn
			}
		} else {
			if isSlow {
				lf = log.ZapWithContext(c).Warn
			}
		}

		lf("http finish log",
			zap.String("method", req.Method),
			zap.String("ip", ip),
			zap.String("user", caller),
			zap.String("path", path),
			zap.String("params", params.Encode()),
			zap.Int("ret", cerr.Code()),
			zap.String("errMsg", cerr.Message()),
			zap.String("stack", fmt.Sprintf("%+v", err)),
			zap.String("err", errmsg),
			zap.Float64("timeout_quota", quota),
			zap.Float64("spendSec", dt.Seconds()),
			zap.String("source", "http-access-log"),
		)
	}
}
