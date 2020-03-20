package resolver

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/discovery"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"google.golang.org/grpc/resolver"
	"net/url"
	"sync"
)

//这里对pkg/discovery/下的驱动进行包装，用于实现grpc内部的resolver和Builder接口，以此使用grpc内部解析机制进行服务发现
const (
	// Scheme is the scheme of discovery address
	Scheme = "grpc"
)

var (
	mu sync.Mutex
)

// 把服务发现驱动注册进来
func Register(b discovery.Builder) {
	mu.Lock()
	defer mu.Unlock()
	if resolver.Get(b.Scheme()) == nil {
		resolver.Register(&Builder{b})
	}
}

// 自己实现的那个Builder并不能和grpc的注册驱动接口兼容，所以在外面包一层，实现grpc的接口方法
type Builder struct {
	discovery.Builder
}

//一样的，包装下discovery.Resolver
// Resolver watches for the updates on the specified target.
// Updates include address updates and service config updates.
type Resolver struct {
	nr   discovery.Resolver
	cc   resolver.ClientConn
	quit chan struct{}
}

//实现grpc内部解析接口
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if len(target.Endpoint) == 0 {
		return nil, errors.Errorf("mygrpc resolver: parse target.Endpoint(%s) failed!err:=endpoint is empty", target.Endpoint)
	}
	r := &Resolver{
		nr:   b.Builder.Build(target.Endpoint),
		cc:   cc,
		quit: make(chan struct{}, 1),
	}
	go r.updateproc() //build的时候直接启动Resolve开始监听服务变化，
	return r, nil
}

func (r *Resolver) updateproc() {
	event := r.nr.Watch()
	for {
		select {
		case <-r.quit:
			return
		case _, ok := <-event:
			if !ok {
				return
			}
		}
		//从内部resolve拉取服务列表，并通过UpdateState接口推送到grpc内部，确保grpc客户端每次拿到的都是最新的服务列表
		if ins, ok := r.nr.Fetch(context.Background()); ok {
			r.newAddress(ins.Instances)
		}
	}
}

func (r *Resolver) newAddress(instances []*discovery.Instance) {
	if len(instances) <= 0 {
		return
	}
	addrs := make([]resolver.Address, 0, len(instances))
	for _, ins := range instances {
		var rpc string
		for _, a := range ins.Addrs {
			u, err := url.Parse(a)
			if err == nil && u.Scheme == Scheme { //筛选出grpc类型的的地址（http类型的地址也在里面，grpc客户端不需要http的）
				rpc = u.Host
			}
		}
		addr := resolver.Address{
			Addr:       rpc,
			ServerName: ins.Name,
		}
		addrs = append(addrs, addr)
	}
	log.SugarWithContext(nil).Debugf("grpc resolver: finally get %d instances:%v", len(addrs), addrs[0].Addr)
	r.cc.UpdateState(resolver.State{Addresses: addrs}) //更新grpc内部服务列表
}

// ResolveNow is a noop for Resolver.
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {
	select {
	case r.quit <- struct{}{}: //关闭updateproc协程
		r.nr.Close() //关闭内部真正的resolve
	default:
	}
}
