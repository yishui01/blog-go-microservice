package internal

import (
	"blog-go-microservice/app/service/poems/internal/dao"
	"blog-go-microservice/app/service/poems/internal/server/grpc"
	"blog-go-microservice/app/service/poems/internal/service"
	"context"
	"github.com/zuiqiangqishao/framework/pkg/log"
	zgrpc "github.com/zuiqiangqishao/framework/pkg/net/grpc"
	"time"
)

type App struct {
	service *service.Service
	grpc    *zgrpc.GrpcServer
}

func InitApp() (app *App, closeFunc func()) {
	d, closeD := dao.New()
	s, closeS, err := service.New(d)
	if err != nil {
		panic("service init err:" + err.Error())
	}
	grpcSrv, _ := grpc.New(s)

	return &App{service: s, grpc: grpcSrv}, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		if err := grpcSrv.Shutdown(ctx); err != nil {
			log.SugarWithContext(ctx).Errorf("grpcSrv Shutdown err(%v)" + err.Error())
		}
		if err := grpcSrv.HttpShutDown(ctx); err != nil {
			log.SugarWithContext(ctx).Errorf("httpSrv Shutdown err(%v)" + err.Error())
		}
		cancel()
		closeD()
		closeS()
	}
}
