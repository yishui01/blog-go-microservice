package khttp

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"github.com/zuiqiangqishao/framework/pkg/trace"
)

//http链路追踪
func Trace() HandlerFunc {
	return func(c *Context) {
		req := c.Request
		spanCtx, err := trace.Tracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			log.SugarWithContext(nil).Debug("Extract trace err", err)
		}
		span := trace.Tracer().StartSpan(req.URL.Path, ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ext.SpanKindRPCServer.Set(span)
		ext.HTTPUrl.Set(span, req.URL.String())
		ext.HTTPMethod.Set(span, req.Method)
		// business tag
		span.SetTag("caller", metadata.String(c.Context, metadata.Caller))

		err = span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			log.SugarWithContext(nil).Error("Inject trace err", err)
		}
		c.Context = opentracing.ContextWithSpan(c.Context, span)
		c.Next()
	}
}
