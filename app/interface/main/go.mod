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
	github.com/gin-gonic/gin v1.6.3
	github.com/jinzhu/gorm v1.9.16
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/qiniu/api.v7 v7.2.5+incompatible
	github.com/qiniu/api.v7/v7 v7.4.1
	github.com/qiniu/x v7.0.8+incompatible // indirect
	github.com/spf13/viper v1.6.2
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/zuiqiangqishao/framework v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.16.0 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1
	qiniupkg.com/x v7.0.8+incompatible // indirect
)
