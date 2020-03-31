package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestDaoGetMetasBySn(t *testing.T) {
	Convey("GetMetasBySn", t, func() {
		var (
			c        = context.Background()
			art      = &model.Article{Content: "TestDaoGetMetasBySn"}
			metas    = &model.Metas{ViewCount: 888}
			notSetSn = "-1sn"
		)
		artId, err := d.CreateArtMetas(context.TODO(), art, metas)
		if err != nil {
			t.Error("create art err", err)
		}
		dbArt, err := d.GetArtFromDB(context.TODO(), "id=?", artId)
		if err != nil {
			t.Error("get art err", err)
		}
		res, err := d.GetMetasBySn(c, dbArt.Sn)
		Convey("Then err ShouldNotBeNil.res ShouldBeNil.", func() {
			So(err, ShouldBeNil)
			So(res.ViewCount, ShouldEqual, metas.ViewCount)
		})
		if err := d.DelCacheMetas(c, notSetSn); err != nil {
			t.Error("delete metas err", err)
		}
		res, err = d.GetMetasBySn(c, notSetSn)
		Convey("notSetSn err ShouldNotBeNil.res ShouldBeNil.", func() {
			So(err, ShouldNotBeNil)
			So(res, ShouldBeNil)
		})
		time.Sleep(time.Millisecond * 200)
		res, err = d.GetMetasBySn(c, notSetSn)
		Convey("notSetSn err ShouldBeNil.res ShouldBeNil.", func() {
			So(err, ShouldBeNil)
			So(res, ShouldBeNil)
		})
	})
}
