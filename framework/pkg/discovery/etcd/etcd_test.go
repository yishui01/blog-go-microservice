package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/zuiqiangqishao/framework/pkg/discovery"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	config := &clientv3.Config{
		Endpoints:   []string{"192.168.136.109:2379", "192.168.136.109:3379", "192.168.136.109:4379"},
		DialTimeout: time.Second * 3,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
	builder, err := New(config)

	if err != nil {
		fmt.Println("etcd 连接失败")
		return
	}
	app1 := builder.Build("app1")

	go func() {
		fmt.Printf("Watch \n")
		for {
			select {
			case <-app1.Watch():
				fmt.Printf("app1 节点发生变化 \n")
			}

		}

	}()
	time.Sleep(time.Second)

	app1Cancel, err := builder.Register(context.Background(), &discovery.Instance{
		Name:     "app1",
		HostName: "h1",
		Addrs:    []string{"120.77.65.70"},
	})
	_, err = builder.Register(context.Background(), &discovery.Instance{
		Name:     "app1",
		HostName: "h1",
		Addrs:    []string{"666.666.666.666"},
	})

	app2Cancel, err := builder.Register(context.Background(), &discovery.Instance{
		Name:     "app2",
		HostName: "h2",
		Addrs:    []string{"000.666.999", "444.555.222.333"},
	})

	if err != nil {
		fmt.Println(err)
	}

	app2 := builder.Build("app2")

	go func() {
		fmt.Println("节点列表")
		for {
			fmt.Printf("app1: ")
			r1, _ := app1.Fetch(context.Background())
			if r1 != nil {
				for _, ins := range r1.Instances {
					fmt.Printf("app: %s host %s,addr:%v \n", ins.Name, ins.HostName, ins.Addrs)
				}
			} else {
				fmt.Printf("\n")
			}
			fmt.Printf("app2: ")
			r2, _ := app2.Fetch(context.Background())
			if r2 != nil {
				for _, ins := range r2.Instances {
					fmt.Printf("app: %s host %s, addr:%v\n", ins.Name, ins.HostName, ins.Addrs)
				}
			} else {
				fmt.Printf("\n")
			}
			time.Sleep(time.Second)
		}
	}()

	time.Sleep(time.Second * 5)
	fmt.Println("取消app1")
	app1Cancel()

	time.Sleep(time.Second * 10)
	fmt.Println("取消app2")
	app2Cancel()

}
