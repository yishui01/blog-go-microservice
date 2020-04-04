package grpc

import (
	"context"
	"flag"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
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
	"strings"
	"sync"
	"time"
)

type GrpcServer struct {
	conf        *ServerConfig
	grpcServer  *grpc.Server
	httpServer  *http.Server
	mutex       sync.RWMutex
	unaryMidle  []grpc.UnaryServerInterceptor
	streamMidle []grpc.StreamServerInterceptor

	//服务注册驱动
	mux        sync.RWMutex
	driverName string
	cancelFunc context.CancelFunc
	register   discovery.Builder
}

func init() {
	addFlag(flag.CommandLine)
	discovery.RegisterDriver(etcd.ETCD_DRIVER_NAME, etcd.GetClosure()) //注册ETCD服务驱动，服务注册和发现都在这个驱动中
}

func addFlag(fs *flag.FlagSet) {
	//这个是mock测试用的，替代真实地址
	fs.Var(&_grpcTarget, "grpc.target", "usage: -grpc.target=seq.service=127.0.0.1:9000 -grpc.target=fav.service=192.168.10.1:9000")
}

// NewServer returns a new blank Server instance with a default server interceptor.
func New(conf *ServerConfig, opt ...grpc.ServerOption) (s *GrpcServer) {
	s = new(GrpcServer)
	if err := s.SetConfig(conf); err != nil {
		panic(errors.Errorf("grpc: set config failed!err: %s", err.Error()))
	}

	keepParam := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     s.conf.IdleTimeout,       //超过该空闲时间就关闭
		MaxConnectionAgeGrace: s.conf.ForceCloseWait,    //关闭时的宽限时间
		Time:                  s.conf.KeepAliveInterval, //触发keepalive的时间
		Timeout:               s.conf.KeepAliveTimeout,  //keepalive 的ack等待时间
		MaxConnectionAge:      s.conf.MaxLifeTime,       //最大生存时间
	})

	//加入一元方法中间件
	//1、recovery
	//2、注册自定义log中间件
	//3、超时时间、ecode转换为grpc code
	//4、jaeger trace
	//5、验证请求参数是否合法
	s.UseUnary(
		s.reovery(),
		grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(trace.Tracer())),
		serverLog(),
		s.handle(),
		s.validate(),
	)

	opt = append(opt,
		keepParam,
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(s.unaryMidle...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(s.streamMidle...)))

	s.grpcServer = grpc.NewServer(opt...)
	return s

}

//启动grpc服务
func (s *GrpcServer) Start() (*GrpcServer, net.Addr, error) {
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
	log.SugarWithContext(nil).Infof("start grpc listen on: %v %v", s.conf.Network, lis.Addr())
	//注册反射服务，可以在运行的时候，通过grpcurl -plaintext localhost:1234 list来看当前端口有哪些服务，
	// 并可直接通过grpcurl工具直接在命令行调用grpc
	reflection.Register(s.grpcServer)
	go func() {
		if err := s.Server().Serve(lis); err != nil {
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
func (s *GrpcServer) SetHttpServer(registerFn func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error), CustomMatcher runtime.HeaderMatcherFunc, opt ...grpc.DialOption) (srv *GrpcServer) {
	if CustomMatcher == nil {
		//设置grpcgateway传递http header的规则，保存到md中供服务端grpc使用（traceId、jwt_user_id等）
		CustomMatcher = func(key string) (k string, bool2 bool) {
			return key, strings.Contains(key, "Uber")
		}
	}

	mux := runtime.NewServeMux(
		//传递header头的匹配规则
		runtime.WithIncomingHeaderMatcher(CustomMatcher),
		//http response 中不要忽略空值字段
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	)

	//这里用的是带Dail的那个注册方法，请求路径为：http=>gateway=>tcp connect=>grpc Server
	err := registerFn(context.Background(), mux, s.conf.Addr, opt)
	if err != nil {
		log.ZapWithContext(nil).Fatal("注册grpc-gateway失败" + err.Error())
	}
	server := &http.Server{
		ReadTimeout:  s.conf.HttpReadTimeout,
		WriteTimeout: s.conf.HttpWriteTimeout,
		Handler:      mux,
		Addr:         s.conf.HttpAddr,
	}
	s.httpServer = server
	return s
}

//将映射后的httpServer启动
func (s *GrpcServer) HttpStart() *GrpcServer {
	if !s.conf.HttpEnable {
		log.ZapWithContext(nil).Info("HTTP Server is not Enable")
		return s
	}
	go func() {
		if s.httpServer == nil {
			panic("http server is not set")
		}
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.ZapWithContext(nil).Fatal("grpcgateway start HttpStart err:" + err.Error())
		}
	}()
	log.SugarWithContext(nil).Infof("start http listen on: %v", s.conf.HttpAddr)
	return s
}

func (s *GrpcServer) HttpStartTLS(certFile, keyFile string) *GrpcServer {
	if !s.conf.HttpEnable {
		log.ZapWithContext(nil).Info("HTTP Server is not Enable")
		return s
	}
	go func() {
		if s.httpServer == nil {
			panic("http server is not set")
		}
		if err := s.httpServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.ZapWithContext(nil).Fatal("grpcgateway start HttpStartTLS err:" + err.Error())
		}
	}()
	log.SugarWithContext(nil).Infof("start https listen on: %v", s.conf.HttpAddr)
	return s
}

//设置各项配置，地址、连接协议、grpc的ctx超时时间、各项keep-alive参数
func (s *GrpcServer) SetConfig(conf *ServerConfig) error {
	if conf == nil {
		conf = _defaultSerConf
	}
	if conf.ServiceName == "" {
		conf.ServiceName = app.GetAppConf().AppName
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

// Shutdown stops the server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished or the context deadline is reached.
func (s *GrpcServer) Shutdown(ctx context.Context) (err error) {
	ch := make(chan struct{})
	s.UnRegister()
	go func() {
		s.grpcServer.GracefulStop()
		close(ch)
	}()

	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
		err = ctx.Err()
	case <-ch:
	}

	return
}

func (s *GrpcServer) HttpShutDown(ctx context.Context) (err error) {
	if s.HttpServer() == nil {
		return errors.New("not have http server")
	}
	return errors.WithStack(s.httpServer.Shutdown(ctx))
}

// Server return the grpc server for registering service.
func (s *GrpcServer) Server() *grpc.Server {
	return s.grpcServer
}

// Server return the grpc server for registering service.
func (s *GrpcServer) HttpServer() *http.Server {
	return s.httpServer
}

// Server return the grpc server for registering service.
func (s *GrpcServer) GetConf() *ServerConfig {
	return s.conf
}
