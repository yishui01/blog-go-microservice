package model

import (
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"gopkg.in/go-playground/validator.v9"
	"time"
)

const USERCATE_BACK = "BACK"
const USERCATE_OPEN = "OPEN"
const USERCATE_EMAIL = "EMAIL"
const USERCATE_PHONE = "PHONE"

type User struct {
	ID            int64      `form:"id" gorm:"column:id" json:"id"` //输入到前台的时候注意要另外建立user结构体隐藏掉ID字段
	Sn            string     `form:"sn" json:"sn"`
	UserName      string     `form:"username" gorm:"column:username" json:"username"`
	PassWord      string     `form:"password" gorm:"column:password" json:"-"`
	PasswordToken string     `form:"password_token" json:"-"`
	NickName      string     `form:"nickname" gorm:"column:nickname" json:"nickname"`
	Avatar        string     `form:"avatar" json:"avatar"`
	Desc          string     `form:"desc" json:"desc"`
	Email         string     `form:"email" json:"email"`
	Phone         string     `form:"phone" json:"phone"`
	Cate          string     `form:"cate" json:"cate"`
	OpenCate      string     `form:"open_cate" gorm:"column:open_cate" json:"open_cate"`
	OpenId        string     `form:"openid" gorm:"column:openid" json:"openid"`
	OpenInfo      string     `form:"openinfo" gorm:"column:openinfo" json:"openinfo"`
	Status        int64      `form:"status" json:"status"`
	ISSuper       int64      `form:"is_super" json:"is_super"`
	CreatedAt     time.Time  `form:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `form:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time `form:"deleted_at" json:"deleted_at"`
}

type LoginForm struct {
	UserName string `json:"username" validate:"required"`
	PassWord string `json:"passwd" validate:"required"`
}

//前台/后台 修改密码都是这个form表单
type UpdatePass struct {
	OldPass          string `form:"old_pass" validate:"required"` //如果不加form标签，就得严格按照字段名大小写来传：OldPass
	NewPass          string `form:"new_pass" validate:"required,eqfield=NewPassConfirmed"`
	NewPassConfirmed string `form:"new_pass_confirmed" validate:"required"`
	Sn               string `form:"sn" validate:"required"`
}

//后台管理user filter
type BackUserQuery struct {
	PageNum   uint   `form:"pageNum" validate:"required,numeric,min=1,max=200000"`
	PageSize  uint   `form:"pageSize" validate:"required,numeric,min=1,max=200000000"`
	UserName  string `form:"username"`
	NickName  string `form:"nickname"`
	Cate      string `form:"cate"`
	OpenCate  string `form:"open_cate"`
	Status    string `form:"status"`
	CreatedAt string `form:"created_at" validate:"omitempty,datetime=2006-01-02"`
	UpdatedAt string `form:"updated_at"  validate:"omitempty,datetime=2006-01-02"`
	ISSuper   string `form:"is_super" validate:"omitempty,numeric,oneof=0 1"`
	ISDelete  string `form:"is_delete" validate:"omitempty,numeric"`
}

//后台管理列表
type BackListUser struct {
	Total    int
	PageNum  uint
	PageSize uint
	Lists    []*User
}

//// 修改密码表单验证自定义错误信息
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
