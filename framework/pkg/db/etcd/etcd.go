package etcd

import (
	"flag"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
	"time"
)

var (
	endpoints          string
	etcdPrefix         string = "micro_srv" //服务注册目录前缀
	username           string
	passwd             string
	leaseTTL           = 60 //秒
	defaultDialTimeout = 5
)

var _conf EtcdConfig

type EtcdConfig struct {
	Endpoints     []string
	Username      string
	Passwd        string
	LeaseTTL      int    //秒
	DialTimeout   int    //客户端连接超时时间
	ServicePrefix string //服务注册前缀
}

func defaultString(env, value string) string {
	v := os.Getenv(env)
	if v == "" {
		return value
	}
	return v
}

func init() {
	// env
	flag.StringVar(&endpoints, "etcd.endpoints", os.Getenv("ETCD_ENDPOINTS"), "etcd.endpoints is etcd endpoints. value: 127.0.0.1:2379,127.0.0.2:2379 etc.")
	flag.StringVar(&etcdPrefix, "etcd.srv_prefix", defaultString("ETCD_PREFIX", etcdPrefix), "etcd globe key prefix or use ETCD_PREFIX env variable. value etcd_prefix etc.")
	flag.StringVar(&username, "etcd.username", os.Getenv("ETCD_USERNAME"), "etcd AUTH username")
	flag.StringVar(&passwd, "etcd.passwd", os.Getenv("ETCD_PASSWD"), "etcd AUTH passwd")

	_conf.Endpoints = strings.Split(endpoints, ",")
	_conf.ServicePrefix = etcdPrefix
	_conf.Username = username
	_conf.Passwd = passwd
	_conf.LeaseTTL = leaseTTL
	_conf.DialTimeout = defaultDialTimeout
}

func Init() {
	if err := viper.Sub("etcd").Unmarshal(&_conf); err != nil {
		log.Fatal("unable to decode _conf config struct, %v", err)
	}
}

func GetConf() EtcdConfig {
	return _conf
}

func GetDefaultClient() (*clientv3.Client, error) {
	c := clientv3.Config{
		Endpoints:   _conf.Endpoints,
		Username:    _conf.Username,
		Password:    _conf.Passwd,
		DialTimeout: time.Duration(_conf.DialTimeout) * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock(), grpc.WithTimeout(1 * time.Second)},
	}
	cli, err := clientv3.New(c)

	if err != nil {
		return nil, errors.Wrap(err, "GetDefaultClient err")
	}
	return cli, nil
}
