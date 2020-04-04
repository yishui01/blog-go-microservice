package dao

import (
	"flag"
	"github.com/zuiqiangqishao/framework/pkg/setting"
	"os"
	"testing"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../test/")
	setting.Init()
	d, _ = New()
	code := m.Run()
	os.Exit(code)
}
