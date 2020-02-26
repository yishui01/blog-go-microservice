package internal

import (
	"blog-go-microservice/app/service/article/internal/dao"
	"blog-go-microservice/app/service/article/internal/server/grpc"
	"blog-go-microservice/app/service/article/internal/service"
	zgrpc "github.com/zuiqiangqishao/framework/pkg/net/grpc"
	"log"
)

type App struct {
	service *service.Service
	//http    *zhttp.HttpEngine
	grpc *zgrpc.GrpcServer
}

func InitApp() (app *App, closeFunc func(), err error) {
	d, cf1 := dao.New()
	s, cf2, err := service.New(d)
	if err != nil {
		log.Fatal("service初始化失败", err)
	}
	//httpSrv, err := http.New(s)
	grpcSrv, err := grpc.New(s)
	return &App{service: s, grpc: grpcSrv}, func() {
		cf1()
		cf2()
	}, err
}

//closeFunc = func() {
//	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
//	if err := g.Shutdown(ctx); err != nil {
//		log.Error("grpcSrv.Shutdown error(%v)", err)
//	}
//	if err := h.Shutdown(ctx); err != nil {
//		log.Error("httpSrv.Shutdown error(%v)", err)
//	}
//	cancel()
//}
