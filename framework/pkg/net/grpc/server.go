package grpc

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/app"
	"github.com/zuiqiangqishao/framework/pkg/discovery"
	"github.com/zuiqiangqishao/framework/pkg/discovery/etcd"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type GrpcServer struct {
	conf        *ServerConfig
	Server      *grpc.Server
	HttpServer  *http.Server
	mutex       sync.RWMutex
	unaryMidle  []grpc.UnaryServerInterceptor
	streamMidle []grpc.StreamServerInterceptor

	//服务注册驱动
	mux        sync.RWMutex
	driverName string
	cancelFunc context.CancelFunc
	builder    discovery.Builder
}

// NewServer returns a new blank Server instance with a default server interceptor.
func New(conf *ServerConfig, opt ...grpc.ServerOption) (s *GrpcServer) {
	s = new(GrpcServer)
	if err := s.SetConfig(conf); err != nil {
		panic(errors.Errorf("grpc: set config failed!err: %s", err.Error()))
	}

	keepParam := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     s.conf.IdleTimeout,       //连接空闲时间，空闲时间达到这个，就关闭连接
		MaxConnectionAge:      s.conf.ForceCloseWait,    //总生命时间，连接存在时间长于这个，关闭
		Time:                  s.conf.KeepAliveInterval, //多少秒内连接没有动静就启动keepalive机制，发送ping包，默认两小时
		Timeout:               s.conf.KeepAliveTimeout,  //每次发送ping包之后的等待超时时间，两次超时直接close
		MaxConnectionAgeGrace: s.conf.MaxLifeTime,       //生命周期到达的时候，关闭连接时给你宽限的最后秒数
	})

	//加入一元方法中间件
	//1、注册自定义log中间件
	//2、注册grpclog
	//3、recovery
	//4、超时时间、ecode转换为grpc code
	//5、jaeger trace
	//6、验证请求参数是否合法
	s.UseUnary(
		serverLog(s.conf.LogFlag),
		grpc_zap.UnaryServerInterceptor(log.ZapLogger),
		s.reovery(),
		s.handle(),
		grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(trace.Tracer)),
		s.validate(),
	)

	opt = append(opt, keepParam,
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(s.unaryMidle...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(s.streamMidle...)))

	s.Server = grpc.NewServer(opt...)
	return s

}

// Start create a new goroutine run server with configured listen addr
// will panic if any error happend
// return server itself
func (s *GrpcServer) Start() (*GrpcServer, net.Addr, error) {
	addr, err := s.startWithAddr()

	if err != nil {
		return nil, nil, err
	}
	return s, addr, nil
}

// StartWithAddr create a new goroutine run server with configured listen addr
// will panic if any error happend
// return server itself and the actually listened address (if configured listen
// port is zero, the os will allocate an unused port)
func (s *GrpcServer) StartWithAddr() (*GrpcServer, net.Addr, error) {
	addr, err := s.startWithAddr()
	if err != nil {
		return nil, nil, err
	}
	return s, addr, nil
}

func (s *GrpcServer) startWithAddr() (net.Addr, error) {
	lis, err := net.Listen(s.conf.Network, s.conf.Addr)
	if err != nil {
		return nil, err
	}
	log.SugarLogger.Infof("start grpc listen on: %v %v", s.conf.Network, lis.Addr())
	//注册反射服务，可以在运行的时候，通过grpcurl -plaintext localhost:1234 list来看当前端口有哪些服务，
	// 并可直接通过grpcurl工具直接在命令行调用grpc
	reflection.Register(s.Server)
	go func() {
		if err := s.Server.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return lis.Addr(), nil
}

//设置一元方法拦截器
func (s *GrpcServer) UseUnary(handlers ...grpc.UnaryServerInterceptor) *GrpcServer {
	finalSize := len(s.unaryMidle) + len(handlers)
	if finalSize >= int(_abortIndex) {
		panic("grpc error: server use too many unary handlers,current length is" + strconv.Itoa(finalSize))
	}
	mergedHandlers := make([]grpc.UnaryServerInterceptor, finalSize)
	copy(mergedHandlers, s.unaryMidle)
	copy(mergedHandlers[len(s.unaryMidle):], handlers)
	s.unaryMidle = mergedHandlers
	return s
}

//流式方法拦截器
func (s *GrpcServer) UseStream(handlers ...grpc.StreamServerInterceptor) *GrpcServer {
	finalSize := len(s.streamMidle) + len(handlers)
	if finalSize >= int(_abortIndex) {
		panic("grpc error: server use too many stream handlers,current length is" + strconv.Itoa(finalSize))
	}
	mergedHandlers := make([]grpc.StreamServerInterceptor, finalSize)
	copy(mergedHandlers, s.streamMidle)
	copy(mergedHandlers[len(s.unaryMidle):], handlers)
	s.streamMidle = mergedHandlers
	return s
}

//用grpcgateway将grpc映射到httpServer
func (s *GrpcServer) GetHttpServer(registerFn func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error), opt ...grpc.DialOption) (srv *http.Server) {
	mux := runtime.NewServeMux()
	t := trace.ClientTracer
	opt = append(opt, grpc.WithUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(t))))
	//这里用的是带Dail的那个注册方法，请求路径为：http=>gateway=>tcp=>grpc Server
	//还有一个注册方法就是直接把Server注册到gateway里，gateway收到请求后直接处理请求，然后返回给前端，不走RPC客户端调用了
	//只是这样的话就不会经过RPC server的中间件链路追踪、log等，这样就无法追踪请求了，所以目前使用的前者
	err := registerFn(context.Background(), mux, s.conf.Addr, opt)
	if err != nil {
		log.ZapLogger.Fatal("注册grpc-gateway失败" + err.Error())
	}
	server := &http.Server{
		ReadTimeout:  s.conf.HttpReadTimeout,
		WriteTimeout: s.conf.HttpWriteTimeout,
		Handler:      mux,
		Addr:         s.conf.HttpAddr,
	}
	s.HttpServer = server
	return server
}

//将映射后的httpServer启动
func (s *GrpcServer) HttpStart(registerFn func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error), opt ...grpc.DialOption) *GrpcServer {
	go func() {
		panic("grpcgateway start http err:" + s.GetHttpServer(registerFn, opt...).ListenAndServe().Error())
	}()
	log.SugarLogger.Infof("start http listen on: %v", s.conf.HttpAddr)
	return s
}

func (s *GrpcServer) HttpStartTLS(registerFn func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error), certFile, keyFile string, opt ...grpc.DialOption) *GrpcServer {
	go func() {
		panic("grpcgateway start http tls err:" + s.GetHttpServer(registerFn, opt...).ListenAndServeTLS(certFile, keyFile).Error())
	}()
	log.SugarLogger.Infof("start https listen on: %v", s.conf.HttpAddr)
	return s
}

//设置各项配置，地址、连接协议、grpc的ctx超时时间、各项keep-alive参数
func (s *GrpcServer) SetConfig(conf *ServerConfig) error {
	discovery.RegisterDriver(etcd.ETCD_DRIVER_NAME, etcd.GetClosure()) //注册ETCD服务发现驱动

	if conf == nil {
		conf = _defaultSerConf
	}

	if conf.Addr == "" {
		conf.Addr = "0.0.0.0:9000"
	}
	if conf.Network == "" {
		conf.Network = "tcp"
	}
	if conf.Timeout <= 0 {
		conf.Timeout = time.Second
	}
	if conf.IdleTimeout <= 0 {
		conf.IdleTimeout = time.Second * 60
	}
	if conf.MaxLifeTime <= 0 {
		conf.MaxLifeTime = time.Hour * 2
	}
	if conf.ForceCloseWait <= 0 {
		conf.ForceCloseWait = time.Second * 20
	}
	if conf.KeepAliveInterval <= 0 {
		conf.KeepAliveInterval = time.Second * 60
	}
	if conf.KeepAliveTimeout <= 0 {
		conf.KeepAliveTimeout = time.Second * 20
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.conf = conf
	return nil

}

//将服务注册到注册中心
func (s *GrpcServer) Register(ins *discovery.Instance) (cancel context.CancelFunc, err error) {
	if s.builder == nil {
		//尝试自动根据配置文件查找驱动
		driverName := viper.GetString("registry.driver")
		f, err := discovery.GetDriver(driverName)
		if err != nil {
			return nil, errors.Wrap(err, "Not set Discovery Builder")
		}

		builder, err := f()
		if err != nil {
			return nil, errors.Wrap(err, "Can not build Discovery Builder")
		}
		s.SetDiscoveryBuilder(driverName, builder)
	}
	if ins == nil {
		ins = &discovery.Instance{
			Name:     app.AppConf.AppName,
			HostName: app.AppConf.HostName,
			Addrs:    []string{s.conf.Addr},
		}
	}
	cancelFunc, err := s.builder.Register(ins)
	s.cancelFunc = cancelFunc
	return cancelFunc, err
}

//反注册服务
func (s *GrpcServer) UnRegister() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}

//设置服务注册驱动
func (s *GrpcServer) SetDiscoveryBuilder(driverName string, build discovery.Builder) *GrpcServer {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.driverName = driverName
	s.builder = build
	return s
}

func (s *GrpcServer) GetDiscoveryDriverName() string {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.driverName
}
func (s *GrpcServer) GetDiscoveryBuilder() discovery.Builder {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.builder
}

// Shutdown stops the server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished or the context deadline is reached.
func (s *GrpcServer) Shutdown(ctx context.Context) (err error) {
	ch := make(chan struct{})
	s.UnRegister()
	go func() {
		s.Server.GracefulStop()
		close(ch)
	}()

	select {
	case <-ctx.Done():
		s.Server.Stop()
		err = ctx.Err()
	case <-ch:
	}

	return
}
