package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

var (
	FileConf *FileConfig = DefaultFileConfig()
)

//日志文件配置
type FileConfig struct {
	Filename   string //log file name
	Level      string //debug  info  warn  error fatal
	Encode     string //json or console
	CallFull   bool   //whether full call path or short path, default is short
	MaxSize    int    //max size of log.(MB)
	MaxAge     int    //time to keep, (day)
	MaxBackups int    //max file numbers
	LocalTime  bool   //(default UTC)
	Compress   bool   //default false
	EnableHost bool   //Add host name and other additional fields
}

func DefaultFileCore() zapcore.Core {
	return FileConf.NewFileCore()
}

func (f *FileConfig) NewFileCore() zapcore.Core {
	if f == nil {
		panic("FileConfig can not be nil")
	}
	if f.Filename == "" {
		return nil
	}

	enCfg := zap.NewProductionEncoderConfig()
	if f.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	enCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	levelEnable := zap.NewAtomicLevelAt(convertLogLevel(f.Level))
	fileEncoder := zapcore.NewJSONEncoder(enCfg)
	if f.Encode == "console" {
		fileEncoder = zapcore.NewConsoleEncoder(enCfg)
	}

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   f.Filename,
		MaxSize:    f.MaxSize,
		MaxAge:     f.MaxAge,
		MaxBackups: f.MaxBackups,
		LocalTime:  f.LocalTime,
	})
	core := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), levelEnable)
	if f.EnableHost {
		core = AddHostInfo(core)
	}

	return core
}

func DefaultFileConfig() *FileConfig {
	return &FileConfig{
		Filename:   "default.log",
		Level:      "info",
		Encode:     "json",
		CallFull:   false,
		MaxSize:    5,     //max size of log.(MB)
		MaxAge:     14,    //time to keep, (day)
		MaxBackups: 30,    //max file numbers
		LocalTime:  true,  //(default UTC)
		Compress:   false, //default false
		EnableHost: true,  //Add host name and other additional fields
	}
}
