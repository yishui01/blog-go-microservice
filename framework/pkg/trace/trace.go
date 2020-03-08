package trace

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/zuiqiangqishao/framework/pkg/app"
	"io"
	"sync"
)

var (
	Tracer opentracing.Tracer
	Closer io.Closer
	once   sync.Once
)

func Init() (opentracing.Tracer, io.Closer) {
	once.Do(func() {
		Tracer, Closer = initJaeger(app.AppConf.AppName, viper.GetString("trace.agentAddr"))
	})

	return Tracer, Closer
}

func initJaeger(service, agentAddr string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler:     &config.SamplerConfig{Type: "const", Param: 1},
		Reporter:    &config.ReporterConfig{LogSpans: true, LocalAgentHostPort: agentAddr},
		ServiceName: service,
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR:cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}
