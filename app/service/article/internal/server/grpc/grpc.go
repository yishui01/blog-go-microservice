package grpc

import (
	pb "blog-go-microservice/app/service/article/api"
	zgrpc "github.com/zuiqiangqishao/framework/pkg/net/grpc"
	"google.golang.org/grpc"
)

func New(svc pb.ArticleServer) (grpcServer *zgrpc.GrpcServer, err error) {
	var conf *zgrpc.ServerConfig
	grpcServer = zgrpc.New(conf)
	pb.RegisterArticleServer(grpcServer.Server, svc)
	grpcServer, _, err = grpcServer.Start() //开启grpc服务
	if err != nil {
		panic("start grpc server err:" + err.Error())
	}
	_, err = grpcServer.Register(nil) //注册到注册中心
	if err != nil {
		panic("Register grpc server err:" + err.Error())
	}
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}
	grpcServer.HttpStart(pb.RegisterArticleHandlerFromEndpoint, opts...) //开启grpc-gateway映射grpc到http

	return
}
