module blog-go-microservice/app/service/article

go 1.14

replace github.com/zuiqiangqishao/framework => ../../../framework

require (
	github.com/douyu/jupiter v0.2.5
	github.com/garyburd/redigo v1.6.2
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.6
	github.com/jinzhu/gorm v1.9.16
	github.com/labstack/echo/v4 v4.1.16
	github.com/olivere/elastic/v7 v7.0.20
	github.com/pkg/errors v0.9.1
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/zuiqiangqishao/framework v0.0.0-00010101000000-000000000000
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20200925080053-05aa5d4ee321 // indirect
	golang.org/x/sys v0.0.0-20200923182605-d9f96fdee20d // indirect
	google.golang.org/genproto v0.0.0-20191216205247-b31c10ee225f
	google.golang.org/grpc v1.26.0

)
