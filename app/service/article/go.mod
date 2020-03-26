module blog-go-microservice/app/service/article

go 1.13

replace github.com/zuiqiangqishao/framework => ../../../framework

require (
	bou.ke/monkey v1.0.2
	github.com/bilibili/kratos v0.3.3 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/g4zhuj/grpc-wrapper v0.0.0-20190508092021-ced55bb6c5d6 // indirect
	github.com/garyburd/redigo v1.6.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.2
	github.com/golang/protobuf v1.3.5
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.13.0
	github.com/jinzhu/gorm v1.9.12
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/olivere/elastic/v7 v7.0.12
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prashantv/gostub v1.0.0
	github.com/prometheus/client_golang v1.4.1
	github.com/segmentio/kafka-go v0.3.5
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/vcraescu/go-paginator v0.0.0-20200304054438-86d84f27c0b3 // indirect
	github.com/zuiqiangqishao/framework v0.0.0-00010101000000-000000000000
	go.etcd.io/etcd v0.0.0-20200320040136-0eee733220fc
	go.uber.org/zap v1.14.1
	golang.org/x/exp/errors v0.0.0-20200221183520-7c80518d1cc7 // indirect
	google.golang.org/genproto v0.0.0-20191216205247-b31c10ee225f
	google.golang.org/grpc v1.26.0
)
