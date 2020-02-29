package grpc

import (
	"github.com/zuiqiangqishao/framework/pkg/app"
	"math"
	"time"
)

var (
	_defaultSerConf = &ServerConfig{
		ServiceName:       app.AppConf.AppName,
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

// ServerConfig is rpc server conf.
type ServerConfig struct {
	ServiceName string
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
