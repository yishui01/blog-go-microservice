package dao

import (
	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic/v7"
	"github.com/zuiqiangqishao/framework/pkg/db/db"
	"github.com/zuiqiangqishao/framework/pkg/db/es"
)

type Dao struct {
	db *gorm.DB
	es *elastic.Client
}

func New() (*Dao, func()) {
	d := &Dao{
		db: NewDB(),
		es: NewESClient(),
	}

	return d, func() { d.Close() }
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}

func NewDB() *gorm.DB {
	return db.NewDB(nil)
}
func NewESClient() *elastic.Client {
	return es.New(nil)
}
