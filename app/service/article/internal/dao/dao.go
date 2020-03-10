package dao

import (
	xredis "github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic/v7"
	"github.com/zuiqiangqishao/framework/pkg/db/db"
	"github.com/zuiqiangqishao/framework/pkg/db/es"
	"github.com/zuiqiangqishao/framework/pkg/db/redis"
	"github.com/zuiqiangqishao/framework/pkg/sync/pipeline/fanout"
)

type Dao struct {
	db         *gorm.DB
	redis      *xredis.Pool
	es         *elastic.Client
	cacheQueue *fanout.Fanout
}

func New() (*Dao, func()) {
	d := &Dao{
		db:         NewDB(),
		redis:      NewRedis(),
		es:         NewESClient(),
		cacheQueue: fanout.New("cache"),
	}
	return d, func() { d.Close() }
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}

func NewRedis() *xredis.Pool {
	return redis.NewRedisPool(nil)
}

func NewDB() *gorm.DB {
	return db.NewDB(nil)
}

func NewESClient() *elastic.Client {
	return es.New(nil)
}
