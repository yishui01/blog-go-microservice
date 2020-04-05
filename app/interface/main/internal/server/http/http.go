package http

import (
	"blog-go-microservice/app/interface/main/internal/service"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
)

var svc *service.Service

func Init(s *service.Service) *khttp.Engine {
	svc = s
	var httpConf = khttp.ServerConfig{}
	if err := viper.Sub("http").Unmarshal(&httpConf); err != nil {
		panic("unable to decode httpConf struct, " + err.Error())
	}
	engine := khttp.DefaultServer(&httpConf)
	initRouter(engine)
	if err := engine.Start(); err != nil {
		panic("engine.Start() Err: " + err.Error())
	}
	return engine
}

func initRouter(e *khttp.Engine) {
	login := svc.Permis.CheckLogin
	admin := svc.Permis.CheckAdmin
	e.POST("/login", svc.UserLogin)
	g := e.Group("/article")
	{
		g.GET("/ping", ping)
		g.GET("/start", login, helloWorld)
		g.GET("/panic", admin, testPanic)
		g.GET("/err", testErr)
	}
}

func testErr(c *khttp.Context) {
	c.JSON([]string{"报错了老弟", "哎哟不错哦"}, ecode.RequestErr)
}

func testPanic(c *khttp.Context) {
	panic("我有点方啊老哥")
}

// example for http request handler.
func helloWorld(c *khttp.Context) {
	log.ZapWithContext(c).Info("测试下完美请求的日志系统")
	c.JSON("helloWorld完美请求", ecode.OK)
}

func ping(c *khttp.Context) {
	c.JSON("pong", nil)
}
