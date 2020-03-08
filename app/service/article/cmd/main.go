package main

import (
	"blog-go-microservice/app/service/article/internal"
	"github.com/zuiqiangqishao/framework/pkg/setting"
)

func main() {
	setting.Init()
	_, closeFunc := internal.InitApp()
	setting.Wait(closeFunc)
}
