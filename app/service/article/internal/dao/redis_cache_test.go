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

	Convey("AddCacheArt", t, func() {
		Convey("When everything goes positive", func() {
			err := d.AddCacheArt(c, val, timeS)
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
		sn    = "testmeta"
		c     = context.Background()
		metas = &model.Metas{Sn: sn, ArticleId: 1, ViewCount: 1, CmCount: 2, LaudCount: 3}
	)

	Convey("SetMetas", t, func() {
		Convey("When everything goes positive", func() {
			err := d.SetCacheMetas(c, metas)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	equal := func(actual interface{}, expected ...interface{}) string {
		a := actual.(map[string]int64)
		e := expected[0].(*model.Metas)
		if e.LaudCount == a[model.LaudRedisKey] &&
			e.ViewCount == a[model.ViewRedisKey] &&
			e.CmCount == a[model.CmRedisKey] &&
			e.ArticleId == a[model.ArtIdRedisKey] {
			return ""
		}

		t.Logf("actual(%#v),expected(%#v)", a, e)
		return "get metas is not equal source metas"
	}
	Convey("GetMetas", t, func() {
		Convey("When everything goes positive", func() {
			p1, err := d.GetCacheMetas(c, sn)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, equal, metas)
			})
		})
	})

	Convey("AddMetas", t, func() {
		Convey("When ViewRedisKey goes positive", func() {
			field := model.ViewRedisKey
			err := d.AddCacheMetas(c, sn, field)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				p1, _ := d.GetCacheMetas(c, sn)
				So(metas.ViewCount+1, ShouldEqual, p1[field])
			})
		})

		Convey("When CmRedisKey goes positive", func() {
			field := model.CmRedisKey
			err := d.AddCacheMetas(c, sn, field)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				p1, _ := d.GetCacheMetas(c, sn)
				So(metas.CmCount+1, ShouldEqual, p1[field])
			})
		})
		Convey("When LaudRedisKey goes positive", func() {
			field := model.LaudRedisKey
			err := d.AddCacheMetas(c, sn, field)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				p1, _ := d.GetCacheMetas(c, sn)
				So(metas.LaudCount+1, ShouldEqual, p1[field])
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
				So(p1, ShouldBeEmpty)
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
