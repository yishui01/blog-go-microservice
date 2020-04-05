package model

import (
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"gopkg.in/go-playground/validator.v9"
	"time"
)

type User struct {
	ID            int64      `json:"-"`
	Sn            string     `json:"sn"`
	UserName      string     `gorm:"column:username" json:"username"`
	PassWord      string     `gorm:"column:password" json:"-"`
	PasswordToken string     `json:"password_token"`
	NickName      string     `json:"nickname"`
	Avatar        string     `json:"avatar"`
	Desc          string     `json:"desc"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	Cate          string     `json:"cate"`
	OpenId        string     `json:"openid"`
	OpenInfo      string     `json:"openinfo"`
	Status        int64      `json:"status"`
	ISSuper       int64      `json:"is_super"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

type UpdatePass struct {
	OldPass          string `form:"old_pass" validate:"required"` //如果不加form标签，就得严格按照字段名大小写来传：OldPass
	NewPass          string `form:"new_pass" validate:"required,eqfield=NewPassConfirmed"`
	NewPassConfirmed string `form:"new_pass_confirmed" validate:"required"`
	Sn               string `form:"sn" validate:"required"`
}

//// 绑定模型获取验证错误的方法
func (r *UpdatePass) GetError(err error) string {
	str := err.Error()

	if val, ok := err.(validator.ValidationErrors); ok {
		s := ""
		t := func(tag string) string {
			switch tag {
			case "required":
				return "不能为空"
			case "eqfield":
				return "两次密码不一致"
			}
			return "错误"
		}
		for _, v := range val {
			f := v.Field()
			switch f {
			case "OldPass":
				f = "旧密码"
			case "NewPass":
				f = "新密码"
			case "NewPassConfirmed":
				f = "确认密码"
			case "Sn":
				f = "用户SN"
			}
			s += f + " " + t(v.Tag()) + ". "
		}
		return s
	}
	return str
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
