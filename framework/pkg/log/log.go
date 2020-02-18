package log

import (
	"flag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"time"
)

var (
	confPath  string
	Config    *LoggerConfig
	ZapLogger *zap.Logger
)

func init() {
	flag.StringVar(&confPath, "conf", "../configs", "default config dir")
}

// Config log config.
type LoggerConfig struct {
	Disable    bool   //disable output into file
	Filename   string //log file name
	Level      string //debug  info  warn  error fatal
	Encoding   string //json or console
	CallFull   bool   //whether full call path or short path, default is short
	MaxSize    int    //max size of log.(MB)
	MaxAge     int    //time to keep, (day)
	MaxBackups int    //max file numbers
	LocalTime  bool   //(default UTC)
	Compress   bool   //default false
	Stdout     bool   //output into stdout default is true
	StdEncode  string //stdout encode
	AppName    string
}

func Init() {
	viper.AddConfigPath(confPath)
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")
	viper.SetConfigName("application")
	if err := viper.ReadInConfig(); err != nil {
		panic("load app config err" + err.Error())
	}
	if err := viper.Sub("log.file").Unmarshal(&Config); err != nil {
		log.Fatalf("unable to decode log config struct, %v", err)
	}

	viper.SetDefault("appname", "unknowAppName")

	Config.AppName = viper.Sub("app").GetString("appname")
	Config.Stdout = viper.Sub("log.stdout").GetBool("enable")
	Config.StdEncode = viper.Sub("log.stdout").GetString("encode")

	InitLogger(Config)
}

func InitLogger(config *LoggerConfig) *zap.Logger {
	enCfg := zap.NewProductionEncoderConfig()
	if config.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	enCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	levelEnable := zap.NewAtomicLevelAt(convertLogLevel(config.Level))

	cores := []zapcore.Core{}

	//file log
	if !config.Disable {
		fileEncoder := zapcore.NewJSONEncoder(enCfg)

		if config.Encoding == "console" {
			fileEncoder = zapcore.NewConsoleEncoder(enCfg)
		}

		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxAge:     config.MaxAge,
			MaxBackups: config.MaxBackups,
			LocalTime:  config.LocalTime,
		})
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), levelEnable))
	}

	//console log
	if config.Stdout {
		consoleEncoder := zapcore.NewJSONEncoder(enCfg)
		if config.StdEncode == "console" {
			consoleEncoder = zapcore.NewConsoleEncoder(enCfg)
		}
		consoleWriter := zapcore.Lock(os.Stdout)
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(consoleWriter), levelEnable))
	}

	allCore := zapcore.NewTee(cores...)

	opts := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(2)}
	ZapLogger = zap.New(allCore, opts...)
	return ZapLogger
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

//NewDefaultLoggerConfig create a default config
func NewDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      "debug",
		Filename:   "./logs",
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 10,
	}
}
