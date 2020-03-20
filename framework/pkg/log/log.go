package log

import (
	"context"
	"github.com/zuiqiangqishao/framework/pkg/app"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

var (
	_handlers []zapcore.Core
	_hmux     sync.Mutex

	_logConf *LogConfig = DefaultLogConf()
	_lmux    sync.RWMutex

	_zapLogger   *zap.Logger
	_sugarLogger *zap.SugaredLogger
)

type _zlogKeyType struct{}
type _slogKeyType struct{}

var _zloggerKey = _zlogKeyType{}
var _sloggerKey = _slogKeyType{}

//通用配置
type LogConfig struct {
	EnableHost bool   //Add host name and other additional fields
	Encode     string //json or console
	Level      string //debug  info  warn  error fatal
	CallFull   bool   //whether full call path or short path, default is short
	Stdout     bool   //default true
}

//默认输出到文件和stdout
func Default() *zap.Logger {
	RegistHandle(DefaultFileCore())
	RegistHandle(DefaultKafkaCore())
	if _logConf != nil && _logConf.Stdout {
		RegistHandle(_logConf.NewStdoutCore())
	}
	return setLogger()
}

//手动构造handler初始化log
func New(core ...zapcore.Core) *zap.Logger {
	_hmux.Lock()
	defer _hmux.Unlock()

	_handlers = core
	return setLogger()
}

//构造zap Logger
func setLogger() *zap.Logger {
	_hmux.Lock()
	defer _hmux.Unlock()

	allCore := zapcore.NewTee(_handlers...)
	opts := []zap.Option{zap.AddCaller()}
	_zapLogger = zap.New(allCore, opts...)
	_sugarLogger = _zapLogger.Sugar()
	return _zapLogger
}

//日志输出handler
func RegistHandle(core ...zapcore.Core) {
	_hmux.Lock()
	defer _hmux.Unlock()

	_handlers = append(_handlers, core...)
}

func AddHostInfo(core zapcore.Core) zapcore.Core {
	fields := []zap.Field{
		zap.String(app.APP_NAME, app.GetAppConf().AppName),
		zap.String(app.HOST_NAME, app.GetAppConf().HostName),
	}
	return core.With(fields)
}

func NewContext(ctx context.Context, fields ...zapcore.Field) context.Context {
	zcore := ZapWithContext(ctx).With(fields...)
	c := context.WithValue(ctx, _zloggerKey, zcore)
	c = context.WithValue(c, _sloggerKey, zcore.Sugar())
	return c
}

func ZapWithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return _zapLogger
	}
	l := ctx.Value(_zloggerKey)
	ctxLogger, ok := l.(*zap.Logger)
	if ok {
		return ctxLogger
	}
	return _zapLogger
}

func SugarWithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return _sugarLogger
	}
	l := ctx.Value(_sloggerKey)
	ctxLogger, ok := l.(*zap.SugaredLogger)
	if ok {
		return ctxLogger
	}
	return _sugarLogger
}

func DefaultLogConf() *LogConfig {
	return &LogConfig{
		EnableHost: true,
		Encode:     "json",
		Level:      "info",
		CallFull:   false,
		Stdout:     true,
	}
}

func convertLogLevel(levelStr string) (level zapcore.Level) {
	switch levelStr {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	}

	return
}

func GetLogConf() *LogConfig {
	_lmux.RLock()
	defer _lmux.RUnlock()
	t := *_logConf
	return &t
}

func SetLogConf(t *LogConfig) {
	_lmux.Lock()
	defer _lmux.Unlock()
	_logConf = t
}
