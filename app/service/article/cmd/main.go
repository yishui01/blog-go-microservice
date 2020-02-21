package main

import (
	"blog-go-microservice/app/service/article/internal"
	"flag"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/setting"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	flag.Parse()
	setting.InitWithLogger()
	_, closeFunc, err := internal.InitApp()
	if err != nil {
		panic("init app err:" + err.Error())
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.ZapLogger.Info("get a signal " + s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.ZapLogger.Info("article service exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}

}
