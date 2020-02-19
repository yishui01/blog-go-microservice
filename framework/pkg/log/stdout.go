package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

func (f *LogConfig) NewStdoutCore() zapcore.Core {
	enCfg := zap.NewProductionEncoderConfig()
	if f.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	enCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	levelEnable := zap.NewAtomicLevelAt(convertLogLevel(f.Level))

	consoleEncoder := zapcore.NewJSONEncoder(enCfg)
	if f.Encode == "console" {
		consoleEncoder = zapcore.NewConsoleEncoder(enCfg)
	}

	core := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), levelEnable)
	if f.EnableHost {
		core = AddHostInfo(core)
	}
	return core
}

func DefaultStdCore() zapcore.Core {
	return LogConf.NewStdoutCore()
}
