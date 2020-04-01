package grpc

import (
	pb "blog-go-microservice/app/service/webinfo/api"
	zgrpc "github.com/zuiqiangqishao/framework/pkg/net/grpc"
	"google.golang.org/grpc"
)

func New(svc pb.WebInfoServer) (grpcServer *zgrpc.GrpcServer, err error) {
	var conf *zgrpc.ServerConfig
	grpcServer = zgrpc.New(conf)
	pb.RegisterWebInfoServer(grpcServer.Server(), svc)
	grpcServer, _, err = grpcServer.Start() //启动grpc服务
	if err != nil {
		panic("start grpc server err:" + err.Error())
	}
	_, err = grpcServer.Register(nil) //注册到注册中心
	if err != nil {
		panic("Register grpc server to Registration Center err:" + err.Error())
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(zgrpc.GrpcStatusToHttpStatus()),
	}
	grpcServer.SetHttpServer(pb.RegisterWebInfoHandlerFromEndpoint, nil, opts...).HttpStart()

	return
}
