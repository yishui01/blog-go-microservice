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
	_tracer opentracing.Tracer
	_closer io.Closer
	_mu     sync.RWMutex
	_once   sync.Once
)

func Init() (opentracing.Tracer, io.Closer) {
	_once.Do(func() {
		_tracer, _closer = initJaeger(app.GetAppConf().AppName, viper.GetString("trace.agentAddr"))
	})
	return _tracer, _closer
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

func Tracer() opentracing.Tracer {
	_mu.RLock()
	defer _mu.RUnlock()
	return _tracer
}

func SetTracer(t opentracing.Tracer) {
	_mu.Lock()
	defer _mu.Unlock()
	_tracer = t
}
