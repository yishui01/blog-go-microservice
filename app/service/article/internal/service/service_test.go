package service

import (
	"blog-go-microservice/app/service/article/internal/dao"
	"flag"
	"github.com/zuiqiangqishao/framework/pkg/setting"
	"os"
	"testing"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	var (
		d = &dao.Dao{}
	)
	flag.Set("conf", "../../test/")
	setting.Init()
	ss, _, err := New(d)
	if err != nil {
		panic("test service start err "+err.Error())
	}
	s = ss
	code := m.Run()
	os.Exit(code)

}
