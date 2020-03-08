package redis

import (
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"time"
)

type RedisConf struct {
	Proto       string
	Addr        string
	Db          int
	Passwd      string
	Idle        int
	Active      int
	IdleTimeout int
	LifeTime    int
	Dial        func() (redis.Conn, error)
}

func NewRedisPool(c *RedisConf) *redis.Pool {
	if c == nil {
		c = setDefaultConf()
	}
	pool := &redis.Pool{
		MaxIdle:         c.Idle,
		MaxActive:       c.Active,
		IdleTimeout:     time.Duration(c.IdleTimeout) * time.Second,
		MaxConnLifetime: time.Duration(c.LifeTime) * time.Hour,
		Dial:            c.Dial,
	}
	return pool
}

func setDefaultConf() *RedisConf {
	c := new(RedisConf)
	if err := viper.Sub("redis").Unmarshal(&c); err != nil {
		panic("unable to decode RedisConfig struct, %v" + err.Error())
	}
	c.Dial = func() (redis.Conn, error) {
		conn, err := redis.Dial(c.Proto, c.Addr)
		if err != nil {
			return nil, err
		}
		if _, err := conn.Do("AUTH", c.Passwd); err != nil {
			conn.Close()
			return nil, err
		}
		if _, err := conn.Do("SELECT", c.Db); err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	}
	return c
}
