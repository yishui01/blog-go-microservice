package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic/v7"
	. "github.com/zuiqiangqishao/framework/pkg/utils"
	"time"
)

type Dao struct {
	db    *gorm.DB
	redis *redis.Pool
	es    *elastic.Client
}

func New() (*Dao, func()) {
	d := &Dao{
		db:    NewDB(),
		redis: NewRedisPool("192.168.136.109:6379"),
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
	d.db.Close()
	d.redis.Close()
}

func NewDB() *gorm.DB {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "mc_" + defaultTableName
	}
	db, err := gorm.Open("mysql", "root:root@/micro_blog?charset=utf8&parseTime=True&loc=Local")
	PanicErr(err)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(true)
	db.SingularTable(true)
	return db
}

func NewRedisPool(addr string) *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
	return pool
}

func NewESClient() *elastic.Client {
	cf := []elastic.ClientOptionFunc{
		elastic.SetURL("http://192.168.136.109:9200"),
		elastic.SetBasicAuth("elastic", "changeme"),
		elastic.SetSniff(false),
	}
	es, err := elastic.NewClient(cf...)
	if err != nil {
		PanicErr(err)
	}
	return es
}
