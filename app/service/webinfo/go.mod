module blog-go-microservice/app/service/webinfo

go 1.14

replace github.com/zuiqiangqishao/framework => ../../../framework

require (
	github.com/garyburd/redigo v1.6.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/grpc-ecosystem/grpc-gateway v1.13.0
	github.com/jinzhu/gorm v1.9.12
	github.com/pkg/errors v0.9.1
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726
	github.com/smartystreets/goconvey v1.6.4
	github.com/zuiqiangqishao/framework v0.0.0-00010101000000-000000000000
	google.golang.org/genproto v0.0.0-20191216205247-b31c10ee225f
	google.golang.org/grpc v1.26.0
)
