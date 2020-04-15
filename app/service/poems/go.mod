module blog-go-microservice/app/service/poems

go 1.14

replace github.com/zuiqiangqishao/framework => ../../../framework

require (
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway v1.13.0
	github.com/jinzhu/gorm v1.9.12
	github.com/olivere/elastic/v7 v7.0.12
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/zuiqiangqishao/framework v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.14.1
	google.golang.org/genproto v0.0.0-20191216205247-b31c10ee225f
	google.golang.org/grpc v1.26.0
)
