package auth

import (
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
)

//修改密码，自己只能修改自己的，管理员除外
func (a *Auth) SelfPass(c *khttp.Context) {
	userSn := c.Request.Form.Get("sn")
	if userSn == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "缺少sn参数"))
		c.Abort()
		return
	}

	user, err := a.checkUser(c, false)
	if err == nil {
		if user.ISSuper == 1 || userSn == user.Sn {
			setUserInfo(c, user)
			return
		}
	}

	c.JSON(nil, err)
	c.Abort()
	return

}
