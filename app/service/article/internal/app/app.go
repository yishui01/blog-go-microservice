package app

import (
	"blog-go-microservice/app/service/article/internal/app/dao"
	"blog-go-microservice/app/service/article/internal/app/server/grpc"
	"blog-go-microservice/app/service/article/internal/app/service/article"
	"context"
	"github.com/zuiqiangqishao/framework/pkg/log"
	zgrpc "github.com/zuiqiangqishao/framework/pkg/net/grpc"
	"time"
)

type App struct {
	service *article.Service
	grpc    *zgrpc.GrpcServer
}

func InitApp() (app *App, closeFunc func()) {
	d, closeD := dao.New()
	s, closeS, err := article.New(d)
	if err != nil {
		panic("service init err:" + err.Error())
	}
	grpcSrv, _ := grpc.New(s)

	return &App{service: s, grpc: grpcSrv}, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		if err := grpcSrv.Shutdown(ctx); err != nil {
			log.SugarWithContext(ctx).Errorf("grpcSrv Shutdown err(%v)" + err.Error())
		}
		if grpcSrv.GetConf().HttpEnable {
			if err := grpcSrv.HttpShutDown(ctx); err != nil {
				log.SugarWithContext(ctx).Errorf("httpSrv Shutdown err(%v)" + err.Error())
			}
		}
		cancel()
		closeD()
		closeS()

	}
}
