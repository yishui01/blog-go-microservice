package service

import (
	"blog-go-microservice/app/interface/main/internal/model"
	"blog-go-microservice/app/interface/main/middleware/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
	"github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"net/http"
	"strconv"
)

//登陆，发放两个token到cookie中，一个JWT一个CSRF
func (s *Service) UserLogin(c *khttp.Context) {
	form := &model.LoginForm{}
	if err := c.MustBind(form); err != nil {
		return
	}
	user, err := s.d.UserLogin(c, form.UserName, form.PassWord)
	if err != nil {
		if ecode.EqualError(ecode.RequestErr, err) {
			c.JSON(nil, ecode.Error(ecode.RequestErr, "账号或密码错误"))
			c.Abort()
			return
		}
		log.SugarWithContext(c).Errorf("s.d.UserLogin Err:(%#v)", err)
		c.JSON(nil, err)
		c.Abort()
		return
	}

	//签发jwt
	s.setCookieJwt(c, user)
}

func (s *Service) UpdatePassWord(c *khttp.Context) {
	p := new(model.UpdatePass)
	if err := c.Bind(p); err != nil {
		c.JSON(p.GetError(err), ecode.RequestErr)
		c.Abort()
		return
	}
	userSn, _ := c.Get(metadata.UserSn)
	sn, ok := userSn.(string)
	if !ok {
		log.SugarWithContext(c).Errorf("s.UpdatePassWord can not find usersn in context,userSn:(%#v)", userSn)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	if val, ok := c.Get(metadata.IsAdmin); !ok || val == 0 {
		//不是管理员就验证下旧密码，管理员直接改了,如果有多个管理员，那就不能这么玩了，否则可以改别的管理员密码，然而总共就我一个管理员╮(╯▽╰)╭
		err := s.d.UserUpdatePass(c, sn, p.OldPass, p.NewPass)
		if err != nil {
			if err == ecode.NothingFound || err == ecode.RequestErr {
				c.JSON(nil, ecode.Error(ecode.RequestErr, "旧密码错误"))
			} else {
				c.JSON(nil, ecode.ServerErr)
			}
			c.Abort()
			return
		}
	}

	user, err := s.d.FindFirstUser(c, "sn=?", sn)
	if err != nil {
		log.SugarWithContext(c).Errorf("s.d.FindFirstUser select Err sn:(%#v),Err(%#+v)", sn, err)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	//重新签发jwt
	s.setCookieJwt(c, user)
}

func (s *Service) setCookieJwt(c *khttp.Context, user *model.User) {
	jCookie, err1 := s.generateTokenCookie(c, user, false)
	csCookie, err2 := s.generateTokenCookie(c, user, true)
	if err1 != nil || err2 != nil {
		log.SugarWithContext(c).Errorf("s.d.UserLogin err1:(%#+v),err2(%#+v)", err1, err2)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}

	c.Writer.Header().Set("set-cookie", jCookie.String())
	c.Writer.Header().Add("set-cookie", csCookie.String())
	c.JSON(0, nil)
}

func (s *Service) generateTokenCookie(c *khttp.Context, user *model.User, isCSRF bool) (*http.Cookie, error) {
	//可以使用 securecookie  https://github.com/gorilla/securecookie 对cookie进行加密
	//jwt
	jwtToken, err := s.jwt.GenerateToken(jwt.MapClaims{
		auth.ClaimUserSn:          user.Sn,
		auth.ClaimUserPasswdToken: user.PasswordToken,
	}, isCSRF)

	if err != nil {
		return nil, errors.Wrap(err, "s. setCookie GenerateToken err")
	}
	cookieName := auth.JWT_NAME
	httpOnly := true
	if isCSRF {
		//目前每个用户csrf令牌是不会变化的，也可以加个后置中间件每次response都动态设置csrfToken，并存储csrfToken到DB中，每次请求比对
		cookieName = auth.CSRF_NAME
		httpOnly = false
	}

	return &http.Cookie{
		Name:     cookieName,
		Value:    jwtToken,
		Path:     "/",
		MaxAge:   s.jwt.TTLSecond,
		HttpOnly: httpOnly,
	}, nil
}

/**********************************************   后台  ************************************************/
func (s *Service) BackUserList(c *khttp.Context) {
	var (
		query = new(model.BackUserQuery)
		users = []*model.User{}
		err   error
	)
	if err := c.MustBind(query); err != nil {
		return
	}
	if users, err = s.d.BackUserList(c, query); err != nil {
		log.SugarWithContext(c).Error("s.BackUserList Err:(%#+v)", err)
		c.JSON(nil, err)
		c.Abort()
		return
	}

	datas := new(model.BackListUser)
	datas.Lists = users
	datas.Total = len(users)
	datas.PageNum = query.PageNum
	datas.PageSize = query.PageSize
	c.JSON(datas, nil)
}

func (s *Service) BackUserCreate(c *khttp.Context) {
	user := new(model.User)
	if err := c.Bind(user); err != nil {
		//ignore err
	}
	user.Cate = model.USERCATE_BACK
	userId, err := s.d.BackUserCreate(c, user, user.ISSuper != 0)
	if err != nil {
		log.SugarWithContext(c).Error("s.BackUserUpdate Err:(%#+v)", err)
		c.JSON(userId, err)
		c.Abort()
		return
	}

	c.JSON(userId, nil)
}

func (s *Service) BackUserUpdate(c *khttp.Context) {
	user := new(model.User)
	c.Bind(user)
	if user.ID == 0 {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "user id can not be empty"))
		c.Abort()
		return
	}

	selfId, exist := c.Get(metadata.UserId)
	if !exist {
		log.SugarWithContext(c).Error("S.BackUserUpdate can not get user_id in context")
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	if selfId.(int64) == user.ID {
		//管理员自己，那自己就不能禁用自己,以及取消自己的超管权限
		user.Status = 0
		user.ISSuper = 1
	}

	userId, err := s.d.BackUserUpdate(c, user)
	if err != nil {
		if err != ecode.NothingFound {
			log.SugarWithContext(c).Error("s.BackUserUpdate Err:(%#+v)", err)
		}
		c.JSON(userId, err)
		c.Abort()
		return
	}

	c.JSON(userId, nil)
}

func (s *Service) BackUserDelete(c *khttp.Context) {
	id := c.Request.Form.Get("id")
	if id == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "ID 不能为空"))
		c.Abort()
		return
	}
	idNum, err := strconv.Atoi(id)
	if err != nil || idNum <= 0 {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "id参数不合法:"+id))
		c.Abort()
		return
	}
	selfId, exist := c.Get(metadata.UserId)
	if !exist {
		log.SugarWithContext(c).Error("S.BackUserUpdate can not get user_id in context")
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}

	if selfId.(int64) == int64(idNum) {
		//管理员自己，自己不能删除自己
		c.JSON(nil, ecode.Error(ecode.RequestErr, "自己不能删除自己"))
		c.Abort()
		return
	}

	err = s.d.BackUserDelete(c, uint(idNum))
	if err != nil {
		if err == ecode.RequestErr {
			c.JSON(1, nil)
			return
		}
		log.SugarWithContext(c).Error("s.BackUserDelete Err(%#+v)", err)
		c.JSON(0, err)
		return
	}
	c.JSON(nil, nil)
}
