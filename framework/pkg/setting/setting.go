package setting

import (
	"flag"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/app"
	"github.com/zuiqiangqishao/framework/pkg/db/etcd"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/trace"
	"go.uber.org/zap"
	std "log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	confPath string
)

func init() {
	flag.StringVar(&confPath, "conf", "../configs", "default config dir")
}

//读取application.toml里面的配置，并解析到对应的结构体内
func Init() *zap.Logger {
	flag.Parse()
	viper.AddConfigPath(confPath)
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")
	viper.SetConfigName("application")
	if err := viper.ReadInConfig(); err != nil {
		panic("load app config err" + err.Error())
	}
	if err := viper.Sub("app").Unmarshal(&app.AppConf); err != nil {
		std.Fatalf("unable to decode appConfig struct, %v", err)
	}

	if err := viper.Sub("log").Unmarshal(&log.LogConf); err != nil {
		std.Fatalf("unable to decode logFile config struct, %v", err)
	}

	if err := viper.Sub("log").Unmarshal(&log.FileConf); err != nil {
		std.Fatalf("unable to decode logFile config struct, %v", err)
	}

	if err := viper.Sub("log.file").Unmarshal(&log.FileConf); err != nil {
		std.Fatalf("unable to decode logFile config struct, %v", err)
	}
	etcd.Init()
	trace.Init()
	return log.Default()
}

func Wait(closeFunc func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.ZapLogger.Info("get a signal " + s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.ZapLogger.Info(app.AppConf.AppName + " service exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
