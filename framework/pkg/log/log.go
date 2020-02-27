package log

import (
	"github.com/zuiqiangqishao/framework/pkg/app"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

var (
	_handlers []zapcore.Core
	mux       sync.Mutex

	LogConf     *LogConfig = DefaultLogConf()
	ZapLogger   *zap.Logger
	SugarLogger *zap.SugaredLogger
)

func init() {
	Default()
}

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
	if LogConf != nil && LogConf.Stdout {
		RegistHandle(LogConf.NewStdoutCore())
	}
	return setLogger()
}

//手动构造handler初始化log
func New(core ...zapcore.Core) *zap.Logger {
	mux.Lock()
	defer mux.Unlock()

	_handlers = core
	return setLogger()
}

//构造zap Logger
func setLogger() *zap.Logger {
	mux.Lock()
	defer mux.Unlock()

	allCore := zapcore.NewTee(_handlers...)
	opts := []zap.Option{zap.AddCaller()}
	ZapLogger = zap.New(allCore, opts...)
	SugarLogger = ZapLogger.Sugar()
	return ZapLogger
}

//日志输出handler
func RegistHandle(core ...zapcore.Core) {
	mux.Lock()
	defer mux.Unlock()

	_handlers = append(_handlers, core...)
}

func AddHostInfo(core zapcore.Core) zapcore.Core {
	fields := []zap.Field{
		zap.String(app.APP_NAME, app.AppConf.AppName),
		zap.String(app.HOST_NAME, app.AppConf.HostName),
	}
	return core.With(fields)
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
