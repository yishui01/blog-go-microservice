package service

import (
	"blog-go-microservice/app/service/article/internal/app/model"
	"blog-go-microservice/app/service/article/internal/app/service/article"
	"blog-go-microservice/app/service/article/internal/app/service/article/impl"
)

var (
	UserRepository article.Repository
)

//Init instantiate the service
func Init() {
	UserRepository = impl.NewMysqlImpl(model.MysqlHandler)
}
