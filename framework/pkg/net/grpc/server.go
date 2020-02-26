package grpc

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode/transform"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	_defaultSerConf = &ServerConfig{
		Network:           "tcp",
		Addr:              "0.0.0.0:9000",
		Timeout:           time.Second,
		IdleTimeout:       time.Second * 60,
		MaxLifeTime:       time.Hour * 2,
		ForceCloseWait:    time.Second * 20,
		KeepAliveInterval: time.Second * 60,
		KeepAliveTimeout:  time.Second * 20,
		HttpAddr:          "0.0.0.0:9001",
		HttpReadTimeout:   time.Second * 3,
		HttpWriteTimeout:  time.Second * 20,
	}
	_abortIndex int8 = math.MaxInt8 / 2
)

type GrpcServer struct {
	conf        *ServerConfig
	Server      *grpc.Server
	HttpServer  *http.Server
	mutex       sync.RWMutex
	unaryMidle  []grpc.UnaryServerInterceptor
	streamMidle []grpc.StreamServerInterceptor
}

// ServerConfig is rpc server conf.
type ServerConfig struct {
	// Network is grpc listen network,default value is tcp
	Network string
	// Addr is grpc listen addr,default value is 0.0.0.0:9000
	Addr string
	// Timeout is context timeout for per rpc call.
	Timeout time.Duration
	// IdleTimeout is a duration for the amount of time after which an idle connection would be closed by sending a GoAway.
	// Idleness duration is defined since the most recent time the number of outstanding RPCs became zero or the connection establishment.
	IdleTimeout time.Duration
	// MaxLifeTime is a duration for the maximum amount of time a connection may exist before it will be closed by sending a GoAway.
	// A random jitter of +/-10% will be added to MaxConnectionAge to spread out connection storms.
	MaxLifeTime time.Duration
	// ForceCloseWait is an additive period after MaxLifeTime after which the connection will be forcibly closed.
	ForceCloseWait time.Duration
	// KeepAliveInterval is after a duration of this time if the server doesn't see any activity it pings the client to see if the transport is still alive.
	KeepAliveInterval time.Duration
	// KeepAliveTimeout  is After having pinged for keepalive check, the server waits for a duration of Timeout and if no activity is seen even after that
	// the connection is closed.
	KeepAliveTimeout time.Duration
	// LogFlag to control log behaviour
	// Disable: 1 DisableArgs: 2 DisableInfo: 4
	LogFlag int8

	//grpc-gateway config
	HttpAddr         string
	HttpReadTimeout  time.Duration
	HttpWriteTimeout time.Duration
}

//处理当前grpc配置的超时时间和上游传下来的ctx剩余时间，设置当前请求的最终超时时间
func (s *GrpcServer) handle() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			cancel func()
		)
		s.mutex.RLock()
		conf := s.conf
		s.mutex.RUnlock()
		//将当前grpc剩余的超时时间和配置的超时时间比较，取较小的那个
		timeout := conf.Timeout
		if de, ok := ctx.Deadline(); ok {
			ctimeout := time.Until(de)
			if ctimeout-time.Millisecond*20 > 0 {
				ctimeout = ctimeout - time.Millisecond*20
			}
			if timeout > ctimeout {
				timeout = ctimeout
			}
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		resp, err = handler(ctx, req)
		return resp, transform.FromError(err).Err()
	}
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

func (s *GrpcServer) HttpStart(registerFn func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error), opt ...grpc.DialOption) {
	go func() {
		panic("grpcgateway start http err:" + s.GetHttpServer(registerFn, opt...).ListenAndServe().Error())
	}()
	log.SugarLogger.Infof("start http listen on: %v", s.conf.HttpAddr)
}

func (s *GrpcServer) HttpStartTLS(registerFn func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error), certFile, keyFile string, opt ...grpc.DialOption) {
	go func() {
		panic("grpcgateway start http tls err:" + s.GetHttpServer(registerFn, opt...).ListenAndServeTLS(certFile, keyFile).Error())
	}()
	log.SugarLogger.Infof("start https listen on: %v", s.conf.HttpAddr)
}

//用grpcgateway将grpc映射到http，这里用带Dail的那个注册方法注册httpHandler，不然不会经过RPC server的中间件链路追踪、log等
func (s *GrpcServer) GetHttpServer(registerFn func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error), opt ...grpc.DialOption) (srv *http.Server) {
	mux := runtime.NewServeMux()
	t := trace.ClientTracer
	opt = append(opt, grpc.WithUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(t))))
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

//设置各项配置，地址、连接协议、grpc的ctx超时时间、各项keep-alive参数
func (s *GrpcServer) SetConfig(conf *ServerConfig) error {

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

// Shutdown stops the server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished or the context deadline is reached.
func (s *GrpcServer) Shutdown(ctx context.Context) (err error) {
	ch := make(chan struct{})
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
