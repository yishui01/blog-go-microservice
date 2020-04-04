package poems_service_v1

import (
	"context"
	"fmt"
	"github.com/zuiqiangqishao/framework/pkg/net/grpc"
	g "google.golang.org/grpc"
)

//const ServerAddr2 = "127.0.0.1:9000"
const ServerName = "poems"

// NewClient new grpc client
func NewClient(cfg *grpc.ClientConfig, opts ...g.DialOption) (PoemsClient, error) {
	client := grpc.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("etcd:///%s", ServerName))
	if err != nil {
		return nil, err
	}
	return NewPoemsClient(cc), nil
}
