package http

import (
	pb "blog-go-microservice/app/service/article/api"
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/net/http"
	xhttp "net/http"
)

var svc pb.ArticleServer

func New(s pb.ArticleServer) (engine *http.HttpEngine, err error) {
	svc = s
	engine = http.DefaultServer(http.DefaultServerConfig())
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *http.HttpEngine) {

	g := e.Engine.Group("/article")
	{
		g.GET("/ping", ping)
		g.GET("/get", art)
		g.GET("/start", helloWorld)
		g.GET("/panic", testPanic)
	}
}

func testErr(c *gin.Context) {
	c.AbortWithError(555, errors.New("报错了老弟"))
}

func testPanic(c *gin.Context) {
	panic("我有点方啊老哥")
}

// example for http request handler.
func helloWorld(c *gin.Context) {
	art := new(model.Article)
	art.Title = "zuiqiangqishao"
	c.JSON(200, art)
}

func ping(ctx *gin.Context) {
	if _, err := svc.Ping(context.TODO(), nil); err != nil {
		log.ZapWithContext(ctx).Info("ping error" + err.Error())
		ctx.AbortWithStatus(xhttp.StatusServiceUnavailable)
	}
}

// example for http request handler.
func art(c *gin.Context) {
	//todo

	c.JSON(200, "还没写grpc")
}
