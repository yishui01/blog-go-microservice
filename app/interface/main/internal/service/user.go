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
)

//登陆，发放两个token到cookie中，一个JWT一个CSRF
func (s *Service) UserLogin(c *khttp.Context) {
	params := c.Request.Form
	username := params.Get("username")
	passwd := params.Get("passwd")
	if username == "" || passwd == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "用户名和密码不能为空"))
		c.Abort()
		return
	}
	user, err := s.d.UserLogin(c, username, passwd)
	if err != nil {
		if ecode.EqualError(ecode.Unauthorized, err) {
			c.JSON(nil, err)
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
