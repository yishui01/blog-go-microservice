package http

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

type HttpEngine struct {
	Engine *gin.Engine
	server atomic.Value // store *http.Server
	conf   *ServerConfig
}

type ServerConfig struct {
	Network      string
	Addr         string
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Network:      "tcp",
		Addr:         ":8080",
		Timeout:      time.Second * 3,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 10,
	}
}

func NewServer(conf *ServerConfig) *HttpEngine {
	return &HttpEngine{Engine: gin.New(), conf: conf}
}

func DefaultServer(conf *ServerConfig) *HttpEngine {
	engine := NewServer(conf)
	engine.Engine.Use(Recovery(), Trace(), Logger())
	return engine
}

// Start listen and serve bm engine by given DSN.
func (engine *HttpEngine) Start() error {
	conf := engine.conf
	l, err := net.Listen(conf.Network, conf.Addr)
	if err != nil {
		err = errors.Wrapf(err, "HTTP server: listen tcp: %s", conf.Addr)
		return err
	}

	log.SugarLogger.Infof("HTTP server: start http listen addr: %s", conf.Addr)
	server := &http.Server{
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}
	go func() {
		if err := engine.RunServer(server, l); err != nil {
			if errors.Cause(err) == http.ErrServerClosed {
				log.ZapLogger.Info("HTTP server: server closed")
				return
			}
			panic(errors.Wrapf(err, "HTTP server: engine.ListenServer(%+v, %+v)", server, l))
		}
	}()

	return nil
}

// RunServer will serve and start listening HTTP requests by given server and listener.
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *HttpEngine) RunServer(server *http.Server, l net.Listener) (err error) {
	server.Handler = engine.Engine
	engine.server.Store(server)
	if err = server.Serve(l); err != nil {
		err = errors.Wrapf(err, "listen server: %+v/%+v", server, l)
		return
	}
	return
}
