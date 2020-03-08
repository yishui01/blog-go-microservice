package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	_ "github.com/go-sql-driver/mysql"
	xredis "github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic/v7"
	"github.com/zuiqiangqishao/framework/pkg/db/db"
	"github.com/zuiqiangqishao/framework/pkg/db/es"
	"github.com/zuiqiangqishao/framework/pkg/db/redis"
	. "github.com/zuiqiangqishao/framework/pkg/utils"
)

type Dao struct {
	db    *gorm.DB
	redis *xredis.Pool
	es    *elastic.Client
}

func New() (*Dao, func()) {
	d := &Dao{
		db:    NewDB(),
		redis: NewRedis(),
		es:    NewESClient(),
	}
	return d, func() { d.Close() }
}

func (d *Dao) GetArtBySn(sn string) (*model.Article, error) {
	art := &model.Article{}
	return art, d.db.Where("sn=?", sn).First(&art).Error
}

func (d *Dao) CreateArt() (*model.Article, error) {
	art := &model.Article{
		Sn:      Md5ByTime("keys"),
		Title:   "第一篇文章",
		Img:     "https://avatars3.githubusercontent.com/u/20850040?s=460&v=4",
		Content: "你好啊，世界",
		Status:  1,
	}
	return art, d.db.Create(&art).Error

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
