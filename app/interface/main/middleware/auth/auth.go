package auth

import (
	"blog-go-microservice/app/interface/main/internal/dao"
	"blog-go-microservice/app/interface/main/internal/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
	"github.com/zuiqiangqishao/framework/pkg/net/metadata"
	"net/url"
	"regexp"
	"strings"
)

var (
	_jwtPlace = map[string]bool{
		"cookie": true,
		"header": true,
	}
)

// Config is the identify config model.
type Config struct {
	EnableCSRF   bool
	AllowHosts   []string
	AllowPattern []string

	JWTPlace string //header、cookie //jwt存放位置
	JWTCfg   *JWTCfg
}

// Auth is the authorization middleware
type Auth struct {
	conf *Config
	dao  *dao.Dao
}

var _defaultConf = &Config{
	EnableCSRF: true,
	JWTPlace:   "cookie", //默认把jwt存cookie里，配合csrf校验
}

// New is used to create an authorization middleware
func New(conf *Config, dao *dao.Dao) *Auth {
	if dao == nil {
		panic("dao can not be empty")
	}
	if conf == nil {
		conf = _defaultConf
		conf.JWTCfg = LoadJWTConfInFile()
	}
	if !_jwtPlace[conf.JWTPlace] {
		panic("jwtplace is invalid:" + conf.JWTPlace)
	}
	auth := &Auth{
		conf: conf,
		dao:  dao,
	}
	return auth
}

// 判断用户是否登陆，以及用户status是否为0（0为正常，1为禁用）
func (a *Auth) CheckLogin(c *khttp.Context) {
	a.handle(c, false)
}

//判断用户是否登陆，是否为管理员
func (a *Auth) CheckAdmin(c *khttp.Context) {
	a.handle(c, true)
}

func (a *Auth) handle(c *khttp.Context, isAdmin bool) {
	user, err := a.checkUser(c, isAdmin)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	setUserInfo(c, user.ID, user.Sn)
}

func (a *Auth) checkUser(c *khttp.Context, isAdmin bool) (*model.User, error) {
	var (
		maps jwt.MapClaims
		err  error
	)
	if a.conf.JWTPlace == "cookie" {
		maps, err = a.CookieAuth(c)
	} else {
		maps, err = a.TokenAuth(c)
	}

	if err != nil {
		return nil, ecode.Unauthorized
	}

	userSn := maps[ClaimUserSn].(string)
	passwdToken := maps[ClaimUserPasswdToken].(string)

	// 这里应该http或者grpc调用远程用户服务判断当前用户是否存在，是否禁用。
	// 本项目用户服务集成在网关层，因此这里用本地user查询代替
	user, err := a.getUserStatusBySn(c, userSn)
	if err != nil {
		if err == ecode.NothingFound {
			return nil, ecode.Unauthorized
		}
		log.SugarWithContext(c).Error("jwt getUserStatusBySn err:(%#+v)", err)
		return nil, ecode.ServerErr
	}
	if user.PasswordToken != passwdToken {
		return nil, ecode.Unauthorized
	}
	if user.Status == 1 {
		return nil, ecode.AccessDenied
	}
	if isAdmin && user.ISSuper == 0 {
		return nil, ecode.AccessDenied
	}
	return user, nil
}

// UserWeb is used to mark path as web access required.
func (a *Auth) CookieAuth(c *khttp.Context) (jwt.MapClaims, error) {
	jCookie, err := c.Request.Cookie(JWT_NAME)
	if err != nil {
		return nil, ecode.Unauthorized
	}
	jwtClaims, err := a.checkToken(jCookie.Value, false)
	if err != nil {
		return nil, ecode.Unauthorized
	}
	if a.conf.EnableCSRF && strings.ToUpper(c.Request.Method) == "POST" {
		if err = a.checkCSRF(c, jwtClaims[ClaimUserSn].(string)); err != nil {
			return nil, ecode.Unauthorized
		}
	}
	return jwtClaims, nil
}

// UserMobile is used to mark path as mobile access required.
func (a *Auth) TokenAuth(c *khttp.Context) (jwt.MapClaims, error) {
	jHeader := c.Request.Header.Get(JWT_NAME)
	return a.checkToken(jHeader, false)
}

func (a *Auth) checkToken(tokenStr string, isCSRF bool) (jwt.MapClaims, error) {
	claims, err := a.conf.JWTCfg.ParseJWTToken(tokenStr, false)
	if err != nil {
		return nil, ecode.Unauthorized
	}

	if _, ok := claims[ClaimUserSn].(string); !ok {
		return nil, ecode.Unauthorized
	}
	if _, ok := claims[ClaimUserPasswdToken].(string); !ok {
		return nil, ecode.Unauthorized
	}

	return claims, nil
}

// set mid into context
// NOTE: This method is not thread safe.
func setUserInfo(ctx *khttp.Context, userId int64, sn string) {
	ctx.Set(metadata.UserSn, sn)
	ctx.Set(metadata.UserId, userId)
	if md, ok := metadata.FromContext(ctx); ok {
		md[metadata.UserSn] = sn
		md[metadata.UserId] = userId
		return
	}
}

func (a *Auth) getUserStatusBySn(c *khttp.Context, userSn string) (*model.User, error) {
	return a.dao.FindFirstUser(c, "sn=?", userSn)
}

// CSRF returns the csrf middleware to prevent invalid cross site request.
// Only referer is checked currently.
func (a *Auth) checkCSRF(c *khttp.Context, usreSn string) error {
	validations := []func(*url.URL) bool{}

	addHostSuffix := func(suffix string) {
		validations = append(validations, matchHostSuffix(suffix))
	}
	addPattern := func(pattern string) {
		validations = append(validations, matchPattern(regexp.MustCompile(pattern)))
	}
	for _, r := range a.conf.AllowHosts {
		addHostSuffix(r)
	}
	for _, p := range a.conf.AllowPattern {
		addPattern(p)
	}
	referer := c.Request.Header.Get("Referer")
	if referer == "" {
		return ecode.Unauthorized
	}

	if uri, err := url.Parse(referer); err == nil && uri.Host != "" {
		for _, validate := range validations {
			if validate(uri) {
				return nil
			}
		}
	}

	//check csrf token
	token := c.Request.Header.Get(CSRF_NAME)
	if token == "" {
		return ecode.Unauthorized
	}
	claims, err := a.conf.JWTCfg.ParseJWTToken(token, true)
	if err != nil {
		return ecode.Unauthorized
	}
	if claims["uid"] != usreSn {
		return ecode.Unauthorized
	}
	return nil
}

func matchHostSuffix(suffix string) func(*url.URL) bool {
	return func(uri *url.URL) bool {
		return strings.HasSuffix(strings.ToLower(uri.Host), suffix)
	}
}

func matchPattern(pattern *regexp.Regexp) func(*url.URL) bool {
	return func(uri *url.URL) bool {
		return pattern.MatchString(strings.ToLower(uri.String()))
	}
}
