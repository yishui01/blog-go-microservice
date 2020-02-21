package http

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/zuiqiangqishao/framework/pkg/trace"
)

//http链路追踪
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		spanCtx, err := trace.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		span := trace.Tracer.StartSpan(req.URL.Path, ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ext.SpanKindRPCServer.Set(span)
		ext.HTTPUrl.Set(span, req.URL.String())
		ext.HTTPMethod.Set(span, req.Method)
		err = span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			panic(err.Error())
		}

		c.Next()
	}
}
