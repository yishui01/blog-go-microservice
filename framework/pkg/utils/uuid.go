package utils

import (
	"crypto/md5"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	"time"
)

func GetUUID() string {
	return uuid.NewV4().String()
}

func Md5(key string) string {
	w := md5.New()
	io.WriteString(w, key)
	md5str := fmt.Sprintf("%x", w.Sum(nil))
	return md5str
}

func SubMd5(key string) string {
	return Md5(key)[:16]
}

func Md5ByTime(key string) string {
	timeTemplate1 := "2006-01-02 15:04:05"
	return Md5(key + time.Unix(time.Now().UnixNano(), 0).Format(timeTemplate1))
}

func SubMd5ByTime(key string) string {
	timeTemplate1 := "2006-01-02 15:04:05"
	return Md5(key + time.Unix(time.Now().UnixNano(), 0).Format(timeTemplate1))[:16]
}
