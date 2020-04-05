module blog-go-microservice/app/interface/main

go 1.14

replace blog-go-microservice/app/service/article => ../../service/article

replace blog-go-microservice/app/service/poems => ../../service/poems

replace blog-go-microservice/app/service/webinfo => ../../service/webinfo

replace github.com/zuiqiangqishao/framework => ../../../framework

require (
	blog-go-microservice/app/service/article v0.0.0-00010101000000-000000000000
	blog-go-microservice/app/service/poems v0.0.0-00010101000000-000000000000
	blog-go-microservice/app/service/webinfo v0.0.0-00010101000000-000000000000
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.5.0
	github.com/jinzhu/gorm v1.9.12
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.6.2
	github.com/zuiqiangqishao/framework v0.0.0-00010101000000-000000000000
	gopkg.in/go-playground/validator.v9 v9.29.1
)
