package main

import (
	"blog-go-microservice/app/interface/main/internal/server/http"
	"blog-go-microservice/app/interface/main/internal/service"
	"github.com/zuiqiangqishao/framework/pkg/setting"
)

//这里是二级网关，在这里进行rpc接口数据聚合、数据格式转换、以及权限控制,对外提供http服务
//理论上应该是每个微服务对应一个二级网关，这样可以相互独立，一个网关挂了不会影响到其他的服务
//但是由于时间原因，暂时将所有网关聚合到这一个服务中，前后台所有请求都会经过这个服务
//用户登录、注册、修改密码的逻辑可以直接放在这一层，因为没有其他的服务需要调用认证逻辑
//在rpc层也可以另外单独开一个用户服务，用于提供认证逻辑之外的服务，比如评论服务获取用户头像
func main() {
	setting.Init()
	svc := service.New()
	http.Init(svc)
	setting.Wait(svc.Close)
}
