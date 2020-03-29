package dao

import (
	"context"
	xredis "github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
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
	jobQueue   *fanout.Fanout
	cronClose  func()
}

func New() (*Dao, func()) {
	d := &Dao{
		db:         NewDB(),
		redis:      NewRedis(),
		es:         NewESClient(),
		cacheQueue: fanout.New("cache", fanout.Worker(5)),
		jobQueue:   fanout.New("job", fanout.Worker(5)),
	}
	go func() {
		d.cronClose = d.CronStart(context.TODO())
	}()
	return d, func() { d.Close() }
}

func (d *Dao) Close() {
	if d.cacheQueue != nil {
		d.cacheQueue.Close()
	}
	if d.jobQueue != nil {
		d.jobQueue.Close()
	}
	if d.cronClose != nil {
		d.cronClose()
	}
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

func (d *Dao) CheckExist(tableName string, query interface{}, args ...interface{}) (bool, error) {
	count := 0
	if err := d.db.Table(tableName).Where(query, args...).Count(&count).Error; err != nil {
		return false, errors.WithStack(err)
	}
	return count > 0, nil
}
