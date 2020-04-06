package http

import (
	"blog-go-microservice/app/interface/main/internal/service"
	"github.com/spf13/viper"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
)

var srv *service.Service

func Init(s *service.Service) *khttp.Engine {
	srv = s
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
	e.POST("/login", srv.UserLogin)
	e.POST("/update/pass", srv.Permis.SelfPass, srv.UpdatePassWord)
	e.GET("/ping", ping)

	//前后台服务的返回数据内容是不一样的，所以肯定是分开

	/*********************前台************************/
	web := e.Group("/home")
	{
		art := web.Group("/article")
		{
			art.GET("/list", srv.HomeArtList)
			art.GET("/detail", srv.HomeArtDetail)
		}

		tag := web.Group("/tag")
		{
			tag.GET("/list", srv.HomeTagList)
		}

		webinfo := web.Group("/webinfo")
		{
			webinfo.GET("/list", srv.HomeWebInfoList)
		}

		poems := web.Group("/poems")
		{
			poems.GET("/list", srv.HomePoemList)
		}
	}

	/********************后台************************/
	back := e.Group("/back", srv.Permis.CheckAdmin)
	{
		art := back.Group("/article")
		{
			art.GET("/list", srv.BackArtList)
			art.GET("/detail", srv.BackArtDetail)
			art.POST("/create", srv.BackArtCreate)
			art.POST("/update", srv.BackArtUpdate)
			art.POST("/delete", srv.BackArtDelete)
		}

		tag := back.Group("/tag")
		{
			tag.GET("/list", srv.BackTagList)
			tag.POST("/create", srv.BackTagCreate)
			tag.POST("/update", srv.BackArtUpdate)
			tag.POST("/delete", srv.BackArtDelete)
		}
		webinfo := back.Group("/webinfo")
		{
			webinfo.GET("/list", srv.BackWebInfoList)
			webinfo.POST("/create", srv.BackWebInfoCreate)
			webinfo.POST("/update", srv.BackWebInfoUpdate)
			webinfo.POST("/delete", srv.BackWebInfoDelete)
		}

		poems := back.Group("/poems")
		{
			poems.GET("/list", srv.BackPoemList)
		}
	}

}

func ping(c *khttp.Context) {
	c.JSON("pong", nil)
}
