package main

import (
	pb "blog-go-microservice/app/service/article/api"
	"context"
	"flag"
	"fmt"
	"github.com/zuiqiangqishao/framework/pkg/setting"
	"google.golang.org/grpc"
	"log"
	"math"
	"time"
)

func main() {
	fmt.Println(int(math.MaxInt64 / time.Hour))
	flag.Parse()
	setting.Init()
	c, err := pb.NewClient(nil, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal("没有连接成功", err)
	}
	time.Sleep(time.Second * 10)
	resp, err := c.GetArtBySn(context.Background(), &pb.ArtDetailRequest{Sn: "666", Status: 1})
	if err != nil {
		log.Println("连接成功，调用失败", err.Error())
	}
	fmt.Printf("调用成功：结果为: %v", resp)

}
