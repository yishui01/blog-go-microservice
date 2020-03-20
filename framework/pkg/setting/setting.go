package setting

import (
	"flag"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/app"
	"github.com/zuiqiangqishao/framework/pkg/db/etcd"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/mq"
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
	af := app.AppConfig{}
	if err := viper.Sub("app").Unmarshal(&af); err != nil {
		std.Fatalf("unable to decode appConfig struct, %v", err)
	}
	app.SetAppConf(&af)

	lf := log.LogConfig{}
	if err := viper.Sub("log").Unmarshal(&lf); err != nil {
		std.Fatalf("unable to decode logFile config struct, %v", err)
	}
	log.SetLogConf(&lf)

	ff := log.FileConfig{}
	if err := viper.Sub("log.file").Unmarshal(&ff); err != nil {
		std.Fatalf("unable to decode logFile config struct, %v", err)
	}
	log.SetFileConf(&ff)

	kf := log.KafkaConfig{}
	if err := viper.Sub("log.kafka").Unmarshal(&kf); err != nil {
		std.Fatalf("unable to decode logKafka config struct, %v", err)
	}
	log.SetKafkaConf(&kf)

	k := mq.KafConf{}
	if err := viper.Sub("kafka").Unmarshal(&k); err != nil {
		std.Fatalf("unable to decode kafka config struct, %v", err)
	}
	mq.SetKafkaConf(&k)

	etcd.Init()
	trace.Init()
	return log.Default()
}

func Wait(closeFunc func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.ZapWithContext(nil).Info("get a signal " + s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.ZapWithContext(nil).Info(app.GetAppConf().AppName + " service exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
