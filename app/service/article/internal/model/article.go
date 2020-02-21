package model

import (
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"time"
)

// UserInfo only contain userid now
type Article struct {
	Aid       int        `gorm:"primary_key";json:"_"`
	Sn        string     `json:"id"`
	Title     string     `json:"title"`
	Img       string     `json:"img"`
	Content   string     `json:"content"`
	Status    int        `json:"status"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (user *Article) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("Sn", utils.Md5ByTime("key"))
	return nil
}
