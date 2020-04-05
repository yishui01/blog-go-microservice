package service

import (
	"blog-go-microservice/app/interface/main/internal/model"
	"blog-go-microservice/app/interface/main/middleware/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
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

func (s *Service) UpdatePassWord(c *khttp.Context) {

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
