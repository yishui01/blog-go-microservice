package main

import (
	pb "blog-go-microservice/app/service/article/pb"
	"context"
	"flag"
	"fmt"
	"github.com/zuiqiangqishao/framework/pkg/setting"
	"google.golang.org/grpc"
	"log"
	"strings"
)

func main() {
	flag.Parse()
	setting.Init()
	c, err := pb.NewClient(nil, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal("没有连接成功", err)
	}
	resp, err := c.GetArtBySn(context.Background(), &pb.ArtDetailRequest{Sn: "666"})
	if err != nil {
		log.Println("连接成功，调用失败", err.Error())
	}
	fmt.Printf("调用成功：结果为: %v", resp)
	strings.Builder{}.String()

}
