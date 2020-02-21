package internal

import (
	"blog-go-microservice/app/service/article/internal/dao"
	"blog-go-microservice/app/service/article/internal/server/http"
	"blog-go-microservice/app/service/article/internal/service"
	fhttp "github.com/zuiqiangqishao/framework/pkg/net/http"
	"log"
)

type App struct {
	service *service.Service
	http    *fhttp.HttpEngine
}

func InitApp() (app *App, closeFunc func(), err error) {
	d, cf1 := dao.New()
	s, cf2, err := service.New(d)
	if err != nil {
		log.Fatal("service初始化失败", err)
	}
	httpSrv, err := http.New(s)
	return &App{service: s, http: httpSrv}, func() {
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
