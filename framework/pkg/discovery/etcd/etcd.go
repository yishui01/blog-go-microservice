package etcd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/discovery"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"google.golang.org/grpc"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	endpoints  string //etcd地址
	etcdPrefix string //服务注册目录前缀

	//秒
	registerTTL        = 90
	defaultDialTimeout = 30
)

var (
	_once    sync.Once
	_builder discovery.Builder
	//ErrDuplication is a register duplication err
	ErrDuplication = errors.New("etcd: instance duplicate registration")
)

func init() {
	addFlag()
}

func addFlag() {
	// env
	flag.StringVar(&endpoints, "etcd.endpoints", os.Getenv("ETCD_ENDPOINTS"), "etcd.endpoints is etcd endpoints. value: 127.0.0.1:2379,127.0.0.2:2379 etc.")
	flag.StringVar(&etcdPrefix, "etcd.prefix", defaultString("ETCD_PREFIX", "micro_etcd_"), "etcd globe key prefix or use ETCD_PREFIX env variable. value etcd_prefix etc.")
}

func defaultString(env, value string) string {
	v := os.Getenv(env)
	if v == "" {
		return value
	}
	return v
}

//etcd服务注册发现总结构
type EtcdBuilder struct {
	cli        *clientv3.Client
	ctx        context.Context
	cancelFunc context.CancelFunc

	mutex    sync.RWMutex        //对map并发写的时候要用锁
	apps     map[string]*appInfo //当前正在 监听/获取 哪些服务
	registry map[string]struct{} //当前注册了那些服务
}

//一个服务名对应一个appInfo
type appInfo struct {
	resolver map[*Resolve]struct{} //这个服务下面拥有的观察者，每一个Resolve就是一个获取服务的对象，他们都在watch服务的变更
	ins      atomic.Value
	e        *EtcdBuilder
	once     sync.Once
}

// Resolve etch resolver.//可以看做是服务消费方
type Resolve struct {
	serviceName string        //要拉取并消费的服务名
	event       chan struct{} //etcd  watch到这个服务的put或者delete时会对着这个ch发消息
	builder     *EtcdBuilder
}

//新建一个EtcdBuilder
func New(c *clientv3.Config) (e *EtcdBuilder, err error) {
	if c == nil {
		if endpoints == "" {
			panic(fmt.Errorf("invalid etcd config endpoints:%+v", endpoints))
		}
		c = &clientv3.Config{
			Endpoints:   strings.Split(endpoints, ","),
			DialTimeout: time.Second * time.Duration(defaultDialTimeout),
			DialOptions: []grpc.DialOption{grpc.WithBlock()},
		}
	}
	cli, err := clientv3.New(*c)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	e = &EtcdBuilder{
		cli:        cli,
		ctx:        ctx,
		cancelFunc: cancel,
		apps:       map[string]*appInfo{},
		registry:   map[string]struct{}{},
	}
	return e, err
}

//创建一个EtcdBuilder，全局单例模式
func Builder(c *clientv3.Config) discovery.Builder {
	_once.Do(func() {
		_builder, _ = New(c)
	})
	return _builder
}

//构建服务resolver，然后可以用它拉取对应的服务节点
//1、传入服务名称，创建一个Resolver，再看下etcdBuild的app里有没有注册过这个servie的appInfo（每个service对应一个appInfo)
//2、没有注册过就注册一个，注册后把新生产的Resolver放入这个appInfo的resolver字段里面的，一个appInfo可以有多个resolver
//3、如果之前注册过这个服务，那就往这个新的Resolve发送一个消息
//4、每个app只要启动一次watch就可以了，所以用once进行app的watch
//5、返回Resolver，以后我们就拿着这个Resolve的Fetch方法来获取服务列表、监听Watch方法返回的ch来监听服务变化。
func (e *EtcdBuilder) Build(serviceName string) discovery.Resolver {
	r := &Resolve{
		serviceName: serviceName,
		builder:     e,
		event:       make(chan struct{}, 1),
	}
	e.mutex.Lock()
	app, ok := e.apps[serviceName]
	if !ok {
		app = &appInfo{
			resolver: make(map[*Resolve]struct{}),
			e:        e,
		}
		e.apps[serviceName] = app
	}
	app.resolver[r] = struct{}{}
	e.mutex.Unlock()
	if ok {
		select {
		case r.event <- struct{}{}:
		default:
		}
	}

	app.once.Do(func() {
		go app.watch(serviceName)
		log.SugarLogger.Infof("etcd: AddWatch(%s) already watch(%v)", serviceName, ok)
	})
	return r
}

// Watch watch instance.
func (r *Resolve) Watch() <-chan struct{} {
	return r.event
}

//就是通过这个方法来获取节点信息的
func (r *Resolve) Fetch(ctx context.Context) (ins *discovery.InstancesInfo, ok bool) {
	r.builder.mutex.RLock()
	app, ok := r.builder.apps[r.serviceName]
	r.builder.mutex.RUnlock()
	if ok {
		ins, ok = app.ins.Load().(*discovery.InstancesInfo)
		return
	}
	return
}

//关闭resolver
func (r *Resolve) Close() error {
	r.builder.mutex.Lock()
	if app, ok := r.builder.apps[r.serviceName]; ok && len(app.resolver) != 0 {
		delete(app.resolver, r)
	}
	r.builder.mutex.Unlock()
	return nil
}

//返回 etcd协议
func (e *EtcdBuilder) Scheme() string {
	return "etcd"

}

//注册服务到ETCD中，并开启goroutine等待反注册
func (e *EtcdBuilder) Register(ctx context.Context, ins *discovery.Instance) (cancelFunc context.CancelFunc, err error) {
	e.mutex.Lock()
	if _, ok := e.registry[ins.Name]; ok {
		err = ErrDuplication
	} else {
		e.registry[ins.Name] = struct{}{}
	}
	e.mutex.Unlock()
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(e.ctx)
	if err = e.register(ctx, ins); err != nil {
		e.mutex.Lock()
		delete(e.registry, ins.Name)
		e.mutex.Unlock()
		cancel()
		return
	}
	ch := make(chan struct{}, 1)
	cancelFunc = context.CancelFunc(func() {
		cancel()
		<-ch
	})

	go func() {

		ticker := time.NewTicker(time.Duration(registerTTL/3) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = e.register(ctx, ins)
			case <-ctx.Done():
				_ = e.unregister(ins)
				ch <- struct{}{}
				return
			}
		}
	}()
	return
}

func (e *EtcdBuilder) Close() error {
	e.cancelFunc()
	return nil
}

//注册和续约公用一个操作
func (e *EtcdBuilder) register(ctx context.Context, ins *discovery.Instance) (err error) {
	prefix := e.keyPrefix(ins)
	val, _ := json.Marshal(ins)

	ttlResp, err := e.cli.Grant(context.TODO(), int64(registerTTL))
	if err != nil {
		log.SugarLogger.Error("etcd: register client.Grant(%v) error(%v)", registerTTL, err)
		return err
	}
	_, err = e.cli.Put(ctx, prefix, string(val), clientv3.WithLease(ttlResp.ID))
	if err != nil {
		log.SugarLogger.Error("etcd: register client.Put(%v) appid(%s) hostname(%s) error(%v)",
			prefix, ins.Name, ins.HostName, err)
		return err
	}
	return nil
}

//反注册
func (e *EtcdBuilder) unregister(ins *discovery.Instance) (err error) {
	prefix := e.keyPrefix(ins)

	if _, err = e.cli.Delete(context.TODO(), prefix); err != nil {
		log.SugarLogger.Error("etcd: unregister client.Delete(%v) appid(%s) hostname(%s) error(%v)",
			prefix, ins.Name, ins.HostName, err)
	}
	log.SugarLogger.Infof("etcd: unregister client.Delete(%v)  appid(%s) hostname(%s) success",
		prefix, ins.Name, ins.HostName)
	return
}

func (e *EtcdBuilder) keyPrefix(ins *discovery.Instance) string {
	return fmt.Sprintf("/%s/%s/%s", etcdPrefix, ins.Name, ins.HostName)
}

/************************************和 appInfo 有关的*********************************************/
//watch这个服务目录下的kv，一旦发生变化并且是更新或者是删除，自动获取最新服务列表并赋值给appInfo.ins字段，并通知appInfo内的所有Resolver
func (a *appInfo) watch(serviceName string) {
	_ = a.fetchStore(serviceName)
	prefix := fmt.Sprintf("/%s/%s/", etcdPrefix, serviceName)
	rch := a.e.cli.Watch(a.e.ctx, prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
				_ = a.fetchStore(serviceName)
			}
		}
	}
}

//获取这个服务目录下的所有kv，存入appInfo的ins（automic.Value)中
func (a *appInfo) fetchStore(serviceName string) error {
	prefix := fmt.Sprintf("/%s/%s/", etcdPrefix, serviceName)
	resp, err := a.e.cli.Get(a.e.ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		log.SugarLogger.Errorf("etcd: fetch client.Get(%s) error(%+v)", prefix, err)
		return err
	}
	ins, err := a.parseIns(resp)
	if err != nil && len(ins.Instances) == 0 {
		return err
	}
	a.store(ins)
	return nil
}

//把刚刚解析好的服务节点信息InstancesInfo存入到appInfo自身的ins字段，然后通知所有的resolver，对着他的event Ch发送信息,发的进就发，发不进不管
func (a *appInfo) store(ins *discovery.InstancesInfo) {
	a.ins.Store(ins)
	a.e.mutex.RLock()
	for rs := range a.resolver {
		select {
		case rs.event <- struct{}{}:
		default:
		}
	}
	a.e.mutex.RUnlock()
}

//获取到某个服务（目录）下的所有节点后，通过该方法将节点解析到InstancesInfo中并返回
func (a *appInfo) parseIns(resp *clientv3.GetResponse) (ins *discovery.InstancesInfo, err error) {
	Ins := &discovery.InstancesInfo{
		Instances: make([]*discovery.Instance, 0),
	}
	var e error
	for _, ev := range resp.Kvs {
		in := new(discovery.Instance)
		err := json.Unmarshal(ev.Value, in)
		if err != nil {
			e = err
			log.ZapLogger.Error("ETCD Parse json instance err:" + err.Error())
			continue
		}
		Ins.Instances = append(Ins.Instances, in)
	}
	return Ins, e
}
