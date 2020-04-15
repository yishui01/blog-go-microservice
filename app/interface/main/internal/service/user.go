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
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/*****************************************   公共    ******************************************************/
//获取用户信息，前后台用这一个
func (s *Service) GetUserInfo(c *khttp.Context) {
	selfId, exist := c.Get(metadata.UserId)
	if !exist {
		log.SugarWithContext(c).Error("S.GetUserInfo can not get user_id in context")
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	user, err := s.d.FindFirstUser(c, "id=?", selfId)
	if err != nil {
		log.SugarWithContext(c).Errorf("s.d.FindFirstUser err:(%#+v)", err)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	userCookie, err := s.setUserInfoCookie(c, user)
	if err != nil {
		log.SugarWithContext(c).Errorf("s.d.GetUserInfo err:(%#+v)", err)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}

	c.Writer.Header().Add("set-cookie", userCookie.String())
	c.JSON(0, nil)
}

//登出，取消cookie，前后台用这一个
func (s *Service) UserLogout(c *khttp.Context) {
	jc := http.Cookie{Name: auth.JWT_NAME, Path: "/", Value: "-1", MaxAge: -1, Expires: time.Now().Add(-100 * time.Hour)}
	csrf := jc
	csrf.Name = auth.CSRF_NAME
	user := jc
	user.Name = auth.USER_INFO_NAME

	c.Writer.Header().Set("set-cookie", (&jc).String())
	c.Writer.Header().Add("set-cookie", (&csrf).String())
	c.Writer.Header().Add("set-cookie", (&user).String())
	c.JSON(0, nil)
}

/***********************************  前台   *************************************************************/
func (s *Service) FrontLogin(c *khttp.Context) {
	s.login(c, false)
}

//登陆，发放两个token到cookie中，一个JWT一个CSRF
func (s *Service) login(c *khttp.Context, checkAdmin bool) {
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
		if ecode.EqualError(ecode.AccessDenied, err) {
			c.JSON(nil, ecode.Error(ecode.RequestErr, "用户已被禁用"))
			c.Abort()
			return
		}
		log.SugarWithContext(c).Errorf("s.d.UserLogin Err:(%#v)", err)
		c.JSON(nil, err)
		c.Abort()
		return
	}

	if checkAdmin && user.ISSuper != 1 {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "账号或密码错误"))
		c.Abort()
		return
	}

	//签发jwt
	s.setCookieJwt(c, user)
}

//前台用户修改自己的密码
func (s *Service) FrontUpdatePassWord(c *khttp.Context) {
	p := new(model.UpdatePass)
	if err := c.MustBind(p); err != nil {
		return
	}
	//获取自己的sn
	userSn, _ := c.Get(metadata.UserSn)
	sn, ok := userSn.(string)
	if !ok {
		log.SugarWithContext(c).Errorf("s.UpdatePassWord can not find usersn in context,userSn:(%#v)", userSn)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}

	//检查旧密码
	err := s.d.CheckOldPass(c, sn, p.OldPass)
	if err != nil {
		if err == ecode.NothingFound || err == ecode.RequestErr {
			c.JSON(nil, ecode.Error(ecode.RequestErr, "旧密码错误"))
		} else {
			c.JSON(nil, ecode.ServerErr)
		}
		c.Abort()
		return
	}
	//更新新密码
	if err := s.d.UserUpdatePass(c, sn, p.NewPass); err != nil {
		log.SugarWithContext(c).Errorf("s.d.UserUpdatePass select Err sn:(%#v),Err(%#+v)", sn, err)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}

	//找出更新后的user
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

/**********************************************   后台  ************************************************/
func (s *Service) BackLogin(c *khttp.Context) {
	s.login(c, true)
}

func (s *Service) BackUpdatePassWord(c *khttp.Context) {
	p := new(model.UpdatePass)
	if err := c.MustBind(p); err != nil {
		return
	}
	userSn, _ := c.Get(metadata.UserSn)
	sn, ok := userSn.(string)
	if !ok {
		log.SugarWithContext(c).Errorf("s.BackUpdatePassWord can not find usersn in context,userSn:(%#v)", userSn)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	//如果是自己改自己的密码，那就要先验证下旧密码
	if userSn == sn {
		err := s.d.CheckOldPass(c, sn, p.OldPass)
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
	//否则不用验证旧密码，直接改
	if err := s.d.UserUpdatePass(c, p.Sn, p.NewPass); err != nil {
		log.SugarWithContext(c).Errorf("s.d.UserUpdatePass select Err sn:(%#v),Err(%#+v)", p.Sn, err)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	//后台修改密码后不签发cookie
}

func (s *Service) BackUserList(c *khttp.Context) {
	var (
		query = new(model.BackUserQuery)
		users = []*model.User{}
		total = 0
		err   error
	)
	if err := c.MustBind(query); err != nil {
		return
	}
	if users, total, err = s.d.BackUserList(c, query); err != nil {
		log.SugarWithContext(c).Error("s.BackUserList Err:(%#+v)", err)
		c.JSON(nil, err)
		c.Abort()
		return
	}

	datas := new(model.BackListUser)
	datas.Lists = users
	datas.Total = total
	datas.PageNum = query.PageNum
	datas.PageSize = query.PageSize
	datas.HiddenPassword()
	c.JSON(datas, nil)
}

func (s *Service) BackUserCreate(c *khttp.Context) {
	user := new(model.User)
	if err := c.MustBind(user); err != nil {
		return
	}
	user.Cate = model.USERCATE_BACK
	userId, err := s.d.BackUserCreate(c, user, user.ISSuper != 0)
	if err != nil {
		if !ecode.EqualError(ecode.UniqueErr, err) {
			log.SugarWithContext(c).Error("s.BackUserUpdate Err:(%#+v)", err)
		} else { //这种业务错误码最后都要转换为标准错误码才行，这样才能映射到http statusCode，客户端只管标准错误码
			err = ecode.Error(ecode.RequestErr, "名称重复")
		}
		c.JSON(userId, err)
		c.Abort()
		return
	}

	c.JSON(userId, nil)
}

func (s *Service) BackUserUpdate(c *khttp.Context) {
	user := new(model.User)
	if err := c.MustBind(user); err != nil {
		return
	}
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
		if _, ok := err.(ecode.Codes); !ok {
			log.SugarWithContext(c).Error("s.BackUserUpdate Err:(%#+v)", err)
		}
		if err == ecode.UniqueErr {
			err = ecode.Error(ecode.RequestErr, "名称重复")
		}
		if err == ecode.NothingFound {
			err = ecode.Error(ecode.RequestErr, "数据不存在")
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

func (s *Service) setCookieJwt(c *khttp.Context, user *model.User) {
	jCookie, err1 := s.generateTokenCookie(c, user, false)
	csCookie, err2 := s.generateTokenCookie(c, user, true)
	userCookie, err3 := s.setUserInfoCookie(c, user)
	if err1 != nil || err2 != nil || err3 != nil {
		log.SugarWithContext(c).Errorf("s.d.UserLogin err1:(%#+v),err2(%#+v)", err1, err2)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}

	c.Writer.Header().Set("set-cookie", jCookie.String())
	c.Writer.Header().Add("set-cookie", csCookie.String())
	c.Writer.Header().Add("set-cookie", userCookie.String())
	c.JSON(0, nil)
}

//将用户个人信息保存到cookie中，方便js提取
func (s *Service) setUserInfoCookie(c *khttp.Context, user *model.User) (*http.Cookie, error) {
	maps := map[string]string{
		"sn":       user.Sn,
		"avatar":   user.Avatar,
		"nickname": user.NickName,
		"desc":     user.Desc,
	}
	var (
		infos []byte
		err   error
	)
	if infos, err = utils.JsonMarshal(maps); err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:   auth.USER_INFO_NAME,
		Value:  url.QueryEscape(string(infos)), //cookie中有些特殊字符不能设置，所以要转码
		Path:   "/",
		MaxAge: s.jwt.TTLSecond,
	}, nil
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
