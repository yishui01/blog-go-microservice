package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoCacheArticle(t *testing.T) {
	var (
		sn    = "testArt"
		c     = context.Background()
		val   = &model.Article{Sn: sn, Id: 123}
		timeS = 0
	)

	Convey("SetCacheArt", t, func() {
		Convey("When everything goes positive", func() {
			err := d.SetCacheArt(c, val, timeS)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("GetCacheArticle", t, func() {
		Convey("When everything goes positive", func() {
			p1, err := d.GetCacheArticle(c, sn)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldResemble, val)
			})
		})
	})

	Convey("DeleteCacheArt", t, func() {
		Convey("When everything goes positive", func() {
			err := d.DeleteCacheArt(c, sn)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
			p1, err := d.GetCacheArticle(c, sn)
			Convey("Then err should be nil.p1 should  be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldBeNil)
			})
		})
	})
}

func TestDaoCacheMetas(t *testing.T) {
	var (
		sn       = "testmeta"
		notSetSn = "not set sn"
		c        = context.Background()
		metas    = &model.Metas{Sn: sn, ArticleId: 1, ViewCount: 1, CmCount: 2, LaudCount: 3}
	)

	Convey("SetCacheMetas", t, func() {
		Convey("When everything goes positive", func() {
			err := d.SetCacheMetas(c, metas, 0)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
			p1, err := d.GetCacheMetas(c, sn)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldResemble, metas)
			})
		})
	})

	Convey("When ViewRedisKey goes positive", t, func() {
		Convey("When ViewRedisKey goes positive", func() {
			field := model.ViewRedisKey
			err := d.IncCacheMetas(c, sn, field)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				p1, _ := d.GetCacheMetas(c, sn)
				So(p1.ViewCount, ShouldEqual, metas.ViewCount+1)
			})

			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				p1, _ := d.GetCacheMetas(c, notSetSn)
				So(p1.ArticleId, ShouldEqual, -1)
				So(p1.Sn, ShouldEqual, notSetSn)
				So(p1.ViewCount, ShouldEqual, 0)
				So(p1.LaudCount, ShouldEqual, 0)
				So(p1.CmCount, ShouldEqual, 0)
			})
		})

		Convey("When CmRedisKey goes positive", func() {
			field := model.CmRedisKey
			err := d.IncCacheMetas(c, sn, field)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				p1, _ := d.GetCacheMetas(c, sn)
				So(p1.CmCount, ShouldEqual, metas.CmCount+1)
			})
		})
		Convey("When LaudRedisKey goes positive", func() {
			field := model.LaudRedisKey
			err := d.IncCacheMetas(c, sn, field)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				p1, _ := d.GetCacheMetas(c, sn)
				So(p1.LaudCount, ShouldEqual, metas.LaudCount+1)
			})
		})
	})

	Convey("DelMetas", t, func() {
		Convey("When everything goes positive", func() {
			err := d.DelCacheMetas(c, sn)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
			p1, err := d.GetCacheMetas(c, sn)
			Convey("Then err should be nil.p1 should  be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldBeNil)
			})
		})
	})
}

func TestDaoCacheKey(t *testing.T) {
	Convey("ArtCacheKey", t, func() {
		var (
			sn = "abc"
		)
		Convey("When everything goes positive", func() {
			p1 := ArtCacheKey(sn)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldEqual, ART_PREFIX+sn)
			})
		})
	})

	Convey("MetasCacheKey", t, func() {
		var (
			sn = "abc"
		)
		Convey("When everything goes positive", func() {
			p1 := MetasCacheKey(sn)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldEqual, META_PREFIX+sn)
			})
		})
	})
}
