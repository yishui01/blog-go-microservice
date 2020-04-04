package dao

import (
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/db/db"
)

type Dao struct {
	db *gorm.DB
}

func New() (*Dao, func()) {
	d := &Dao{
		db: NewDB(),
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
