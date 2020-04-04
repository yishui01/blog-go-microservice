package grpc

import (
	pb "blog-go-microservice/app/service/poems/api"
	zgrpc "github.com/zuiqiangqishao/framework/pkg/net/grpc"
	"google.golang.org/grpc"
)

func New(svc pb.PoemsServer) (grpcServer *zgrpc.GrpcServer, err error) {
	grpcServer = zgrpc.New(zgrpc.GetFileConfig())
	pb.RegisterPoemsServer(grpcServer.Server(), svc)
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
	grpcServer.SetHttpServer(pb.RegisterPoemsHandlerFromEndpoint, nil, opts...).HttpStart()

	return
}
