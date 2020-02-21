package trace

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
)

var Tracer opentracing.Tracer

func InitTracer(service string) (opentracing.Tracer, io.Closer) {
	tracer, closer := initJaeger(service)
	Tracer = tracer
	return tracer, closer
}

func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler:     &config.SamplerConfig{Type: "const", Param: 1},
		Reporter:    &config.ReporterConfig{LogSpans: true, LocalAgentHostPort: "192.168.136.109:6831"},
		ServiceName: service,
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR:cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}
