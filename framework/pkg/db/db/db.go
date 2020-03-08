package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/utils"

	"time"
)

type DBConf struct {
	Driver       string
	Dsn          string
	Active       int
	Idle         int
	IdleTimeout  int
	LogMode      bool
	STable       bool
	TablePrefix  string
	TableHandler func(db *gorm.DB, defaultTableName string) string

	Args []interface{}
}

func NewDB(c *DBConf) *gorm.DB {
	if c == nil {
		c = setDefaultConf()
	}

	db, err := gorm.Open(c.Driver, c.Args...)
	utils.PanicErr(err)

	db.DB().SetMaxOpenConns(c.Active)
	db.DB().SetMaxIdleConns(c.Idle)
	db.DB().SetConnMaxLifetime(time.Duration(c.IdleTimeout) * time.Hour)
	db.LogMode(c.LogMode)
	db.SingularTable(c.STable)
	return db
}

func setDefaultConf() *DBConf {
	var conf = new(DBConf)
	if err := viper.Sub("db").Unmarshal(&conf); err != nil {
		panic("unable to decode DBConfig struct, %v" + err.Error())
	}
	conf.TableHandler = func(db *gorm.DB, defaultTableName string) string {
		return conf.TablePrefix + defaultTableName
	}
	conf.Args = []interface{}{conf.Dsn}
	return conf
}
