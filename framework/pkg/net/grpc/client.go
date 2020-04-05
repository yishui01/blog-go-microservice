package grpc

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/discovery"
	"github.com/zuiqiangqishao/framework/pkg/net/grpc/resolver"
	"github.com/zuiqiangqishao/framework/pkg/setting/flagvar"
	"github.com/zuiqiangqishao/framework/pkg/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	_grpcTarget flagvar.StringVars

	_once           sync.Once
	_defaultCliConf = &ClientConfig{
		Dial:              time.Second * 10,
		Timeout:           time.Millisecond * 300,
		KeepAliveInterval: time.Second * 60,
		KeepAliveTimeout:  time.Second * 20,
	}
	_defaultClient *Client
)

// ClientConfig is rpc client conf.
type ClientConfig struct {
	Dial                   time.Duration
	Timeout                time.Duration
	Method                 map[string]*ClientConfig
	NonBlock               bool
	KeepAliveInterval      time.Duration
	KeepAliveTimeout       time.Duration
	KeepAliveWithoutStream bool
}

// Client is the framework's client side instance, it contains the ctx, opt and interceptors.
// Create an instance of Client, by using NewClient().
type Client struct {
	conf  *ClientConfig
	mutex sync.RWMutex

	opts        []grpc.DialOption
	unaryMidle  []grpc.UnaryClientInterceptor
	streamMidle []grpc.StreamClientInterceptor
}

// Register direct resolver by default to handle direct:// scheme.
func init() {
	//todo...
}

// TimeoutCallOption timeout option.
type TimeoutCallOption struct {
	*grpc.EmptyCallOption
	Timeout time.Duration
}

// WithTimeoutCallOption can override the timeout in ctx and the timeout in the configuration file
func WithTimeoutCallOption(timeout time.Duration) *TimeoutCallOption {
	return &TimeoutCallOption{&grpc.EmptyCallOption{}, timeout}
}

// DefaultClient returns a new default Client instance with a default client interceptor and default dialoption.
// opt can be used to add grpc dial options.
func DefaultClient() *Client {
	_once.Do(func() {
		_defaultClient = NewClient(nil)
	})
	return _defaultClient
}

// NewClient returns a new blank Client instance with a default client interceptor.
// opt can be used to add grpc dial options.
func NewClient(conf *ClientConfig, opt ...grpc.DialOption) *Client {
	SetResolver()
	c := new(Client)
	if err := c.SetConfig(conf); err != nil {
		panic(err)
	}
	c.UseOpt(grpc.WithBalancerName(roundrobin.Name)) //指定客户端轮询负载均衡，以后可扩展为根据服务器当前负载进行选择
	c.UseOpt(opt...)
	return c
}

func SetResolver() {
	driverName := viper.GetString("registry.driver")

	f, err := discovery.GetDriver(driverName)
	if err != nil {
		panic("get discovery driver" + driverName + " err:" + err.Error())
	}
	b, err := f()
	if err != nil {
		panic("grpc client get " + driverName + " builder register err:" + err.Error())
	}
	resolver.Register(b) //注册到grpc的那个resolver中，这下客户端就能直接解析协议了
}

// Dial creates a client connection to the given target.
// Target format is scheme://authority/endpoint?query_arg=value
// example: discovery://default/account.account.service?cluster=shfy01&cluster=shfy02
func (c *Client) Dial(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())
	return c.dial(ctx, target, opts...)
}

// DialTLS creates a client connection over tls transport to the given target.
func (c *Client) DialTLS(ctx context.Context, target string, file string, name string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	var creds credentials.TransportCredentials
	creds, err = credentials.NewClientTLSFromFile(file, name)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return c.dial(ctx, target, opts...)
}

func (c *Client) cloneOpts() []grpc.DialOption {
	dialOptions := make([]grpc.DialOption, len(c.opts))
	copy(dialOptions, c.opts)
	return dialOptions
}

func (c *Client) dial(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	dialOptions := c.cloneOpts()
	if !c.conf.NonBlock {
		dialOptions = append(dialOptions, grpc.WithBlock())
	}
	dialOptions = append(dialOptions, opts...)

	// init default handler
	var handlers []grpc.UnaryClientInterceptor
	handlers = append(handlers, c.recovery())
	handlers = append(handlers, clientLogging())
	handlers = append(handlers, grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(trace.Tracer())))
	handlers = append(handlers, c.unaryMidle...)
	// NOTE: c.handle must be a last interceptor.
	handlers = append(handlers, c.handle())

	dialOptions = append(dialOptions, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(handlers...)))
	c.mutex.RLock()
	conf := c.conf
	c.mutex.RUnlock()
	if conf.Dial > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, conf.Dial)
		defer cancel()
	}

	//看下有没有指定 命令行参数 -grpc ，设置了，并且本次客户端调用的服务就是命令行中的service，那就替换本次服务调用方式为直连，地址为命令行中的地址
	//一般是用于mock测试时使用
	if u, e := url.Parse(target); e == nil {
		v := u.Query()
		u.RawQuery = v.Encode()
		// 比较_grpcTarget中的appid是否等于u.path中的appid，并替换成mock的地址
		for _, t := range _grpcTarget {
			strs := strings.SplitN(t, "=", 2)
			if len(strs) == 2 && ("/"+strs[0]) == u.Path {
				u.Path = "/" + strs[1]
				u.Scheme = "passthrough"
				u.RawQuery = ""
				break
			}
		}
		target = u.String()
	}
	if conn, err = grpc.DialContext(ctx, target, dialOptions...); err != nil {
		fmt.Fprintf(os.Stderr, "mygrpc client: dial %s error %v!", target, err)
	}

	err = errors.WithStack(err)
	return
}

// Use attachs a global inteceptor to the Client.
// For example, this is the right place for a circuit breaker or error management inteceptor.
func (c *Client) UseUnary(handlers ...grpc.UnaryClientInterceptor) *Client {
	finalSize := len(c.unaryMidle) + len(handlers)
	if finalSize >= int(_abortIndex) {
		panic("grpc: client use too many unary handlers")
	}
	mergedHandlers := make([]grpc.UnaryClientInterceptor, finalSize)
	copy(mergedHandlers, c.unaryMidle)
	copy(mergedHandlers[len(c.unaryMidle):], handlers)
	c.unaryMidle = mergedHandlers
	return c
}

// UseOpt attachs a global grpc DialOption to the Client.
func (c *Client) UseOpt(opts ...grpc.DialOption) *Client {
	c.opts = append(c.opts, opts...)
	return c
}

// SetConfig hot reloads client config
func (c *Client) SetConfig(conf *ClientConfig) (err error) {
	if conf == nil {
		conf = _defaultCliConf
	}
	if conf.Dial <= 0 {
		conf.Dial = time.Second * 10
	}
	if conf.Timeout <= 0 {
		conf.Timeout = time.Millisecond * 300
	}
	if conf.KeepAliveInterval <= 0 {
		conf.KeepAliveInterval = time.Second * 60
	}
	if conf.KeepAliveTimeout <= 0 {
		conf.KeepAliveTimeout = time.Second * 20
	}
	c.mutex.Lock()
	c.conf = conf
	c.mutex.Unlock()
	return nil
}
