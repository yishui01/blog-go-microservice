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
	e.POST("/logout", srv.UserLogout)
	e.GET("/ping", ping)
	e.POST("/upload", srv.Permis.CheckLogin, srv.Upload)
	e.GET("/user_info", srv.Permis.CheckLogin, srv.GetUserInfo)

	//前后台服务的返回数据内容是不一样的，所以肯定是分开
	/*********************前台************************/
	e.POST("/home/login", srv.FrontLogin)
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

		user := web.Group("/user", srv.FrontLogin)
		{
			user.POST("/update/pass", srv.FrontUpdatePassWord)
		}

	}

	/********************后台************************/
	e.POST("/back/login", srv.BackLogin)
	back := e.Group("/back", srv.Permis.CheckAdmin)
	{
		back.POST("/update/pass", srv.BackUpdatePassWord)
		back.POST("/user_info", srv.GetUserInfo)

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
			tag.POST("/update", srv.BackTagUpdate)
			tag.POST("/delete", srv.BackTagDelete)
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

		user := back.Group("/user")
		{
			user.GET("/list", srv.BackUserList)
			user.POST("/create", srv.BackUserCreate)
			user.POST("/update", srv.BackUserUpdate)
			user.POST("/delete", srv.BackUserDelete)
		}

		upload := back.Group("/upload")
		{
			upload.GET("/list", srv.UploadList)
			upload.POST("/delete", srv.UploadDelete)
		}
	}

}

func ping(c *khttp.Context) {
	c.JSON("pong", nil)
}
