package dao

import (
	"context"
	"testing"
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
