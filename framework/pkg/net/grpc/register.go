package grpc

import (
	"context"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/app"
	"github.com/zuiqiangqishao/framework/pkg/discovery"
)

//将服务注册到注册中心
func (s *GrpcServer) Register(ins *discovery.Instance) (cancel context.CancelFunc, err error) {
	if s.register == nil {
		//尝试自动根据配置文件查找注册驱动
		driverName := viper.GetString("registry.driver")
		f, err := discovery.GetDriver(driverName)
		if err != nil {
			return nil, errors.Wrap(err, "Not set Registry Builder")
		}

		builder, err := f()
		if err != nil {
			return nil, errors.Wrap(err, "Can not build Registry Builder")
		}
		s.SetRegisterDriver(driverName, builder)
	}
	if ins == nil {
		ins = &discovery.Instance{
			Name:     app.GetAppConf().AppName,
			HostName: app.GetAppConf().HostName,
			Addrs:    []string{"grpc://" + s.conf.Addr},
		}
	}
	cancelFunc, err := s.register.Register(ins)
	s.cancelFunc = cancelFunc
	return cancelFunc, err
}

//反注册服务
func (s *GrpcServer) UnRegister() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}

//设置当前GrpcServer的服务注册驱动
func (s *GrpcServer) SetRegisterDriver(driverName string, build discovery.Builder) *GrpcServer {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.driverName = driverName
	s.register = build
	return s
}

func (s *GrpcServer) GetRegisterDriverName() string {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.driverName
}
func (s *GrpcServer) GetRegister() discovery.Builder {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.register
}
