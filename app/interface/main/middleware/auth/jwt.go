package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"time"
)

const JWT_NAME = "access_token"
const CSRF_NAME = "_csrf"

const ClaimUserSn = "user_sn"
const ClaimUserPasswdToken = "user_unique"

type JWTCfg struct {
	JWTSecret      string
	jwtSecretBytes []byte

	CSRFSecret      string
	csrfSecretBytes []byte

	SignMethod string
	signHMAC   *jwt.SigningMethodHMAC
	TTL        time.Duration
	TTLSecond  int //秒数，用于设置 cookie max-age 以及 jwt过期时间
}

func LoadJWTConfInFile() *JWTCfg {
	var (
		j    = new(JWTCfg)
		hmac *jwt.SigningMethodHMAC
		err  error
	)
	if err = viper.Sub("jwt").Unmarshal(&j); err != nil {
		panic("unable to decode jwt conf")
	}
	if j.JWTSecret == "" || j.SignMethod == "" || j.TTL <= 0 {
		panic(fmt.Sprintf("invalid jwt conf:Secret:(%#v),SignMethod(%#v),TTL(%#v)", j.JWTSecret, j.SignMethod, j.TTL))
	}

	j.TTLSecond = int(j.TTL / time.Second)
	if hmac, err = InitSigningMethod(j.SignMethod); err != nil {
		panic("invalid jwt SignMethod:" + j.SignMethod)
	}

	j.jwtSecretBytes = []byte(j.JWTSecret)
	if j.CSRFSecret != "" {
		j.csrfSecretBytes = []byte(j.CSRFSecret)
	}
	j.signHMAC = hmac

	return j
}

func InitSigningMethod(m string) (*jwt.SigningMethodHMAC, error) {
	switch m {
	case "HS256":
		return jwt.SigningMethodHS256, nil
	case "HS384":
		return jwt.SigningMethodHS384, nil
	case "HS512":
		return jwt.SigningMethodHS512, nil
	}
	return nil, fmt.Errorf("unsupported method: %s", m)
}

func (j *JWTCfg) ParseJWTToken(tokenString string, isCSRF bool) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.WithStack(fmt.Errorf("Unexpected signing method: %v", token.Header["alg"]))
		}
		if isCSRF {
			return j.csrfSecretBytes, nil
		}
		return j.jwtSecretBytes, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid jwt token")
}

func (j *JWTCfg) GenerateToken(claims jwt.MapClaims, isCSRF bool) (string, error) {
	var (
		res string
		err error
	)
	claims["exp"] = time.Now().Add(j.TTL).Unix()
	claims["iat"] = time.Now().Unix()
	token := jwt.NewWithClaims(j.signHMAC, claims)
	secret := j.jwtSecretBytes
	msg := "jwt token signed Err"
	if isCSRF {
		if len(j.jwtSecretBytes) == 0 {
			return "", errors.New("csrf secret is empty")
		}
		secret = j.csrfSecretBytes
		msg = "csrf token signed Err"
	}
	if res, err = token.SignedString(secret); err != nil {
		return "", errors.Wrap(err, msg)
	}
	return res, nil
}
