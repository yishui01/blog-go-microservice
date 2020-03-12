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

	opt := []config.Option{}
	if viper.GetBool("trace.stdout") {
		opt = append(opt, config.Logger(jaeger.StdLogger))
	}
	tracer, closer, err := cfg.NewTracer(opt...)
	if err != nil {
		panic(fmt.Sprintf("ERROR:cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func GetTracer(serviceName string) {

}
