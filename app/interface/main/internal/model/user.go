package model

import (
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"time"
)

type User struct {
	ID            int64
	Sn            string
	UserName      string
	PassWord      string
	PasswordToken string
	NickName      string
	Avatar        string
	Desc          string
	Email         string
	Phone         string
	Cate          string
	OpenId        string
	OpenInfo      string
	Status        int64
	ISSuper       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	err := scope.SetColumn("Sn", utils.SubMd5(utils.GetUUID()))
	if err != nil {
		return err
	}
	err = scope.SetColumn("password_token", utils.GetUUID())
	if err != nil {
		return err
	}
	return nil
}
