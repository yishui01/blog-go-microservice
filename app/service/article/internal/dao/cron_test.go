package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoCronStart(t *testing.T) {
	Convey("CronStart", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			close := d.CronStart(c)
			if close == nil {
				t.Error("close func is nil")
			}
		})
	})
}
