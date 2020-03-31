package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoEsPutArtMetas(t *testing.T) {
	var (
		ctx   = context.Background()
		art   = &model.Article{Id: 777, Content: "asdf", Title: "aaaaaddsss"}
		metas = &model.Metas{ViewCount: 99999}

		req = &model.ArtQueryReq{
			PageRequest: utils.PageRequest{PageNum: 1, PageSize: 2},
			Terms:       "id," + strconv.FormatInt(art.Id, 10),
			Status:      0,
			Order:       "created_at|desc",
			//CreatedAt:   15000000,
			//UpdatedAt:   15000000,
			Unscoped: true,
		}
	)
	Convey("es Operations art ", t, func() {
		//增
		p1, err := d.EsPutArtMetas(ctx, art, metas)
		Convey("EsPutArtMetas err should be nil.p1 should not be nil.", func() {
			So(err, ShouldBeNil)
			So(p1, ShouldNotBeNil)
		})
		So(err, ShouldBeNil)
		So(p1, ShouldNotBeNil)

		time.Sleep(time.Second * 3)
		//查
		p1s, err := d.EsSearchArtMetas(ctx, req)
		So(err, ShouldBeNil)
		So(p1s, ShouldNotBeNil)

		arts, err := d.ArtMetasList(ctx, req)
		So(len(arts), ShouldEqual, 1)
		So(arts[0].Content, ShouldEqual, art.Content)
		So(arts[0].Title, ShouldEqual, art.Title)
		So(arts[0].ViewCount, ShouldEqual, metas.ViewCount)

		//改
		newArt := &model.Article{Id: art.Id, Content: "aaaaaaaa"}
		newMetas := &model.Metas{ViewCount: 66666666}
		p2, err := d.EsUpdateArtMetas(ctx, newArt, newMetas)
		So(err, ShouldBeNil)
		So(p2, ShouldNotBeNil)

		//改完要等会查
		time.Sleep(time.Second * 2)
		narts, err := d.ArtMetasList(ctx, req)
		So(len(narts), ShouldEqual, 1)
		So(narts[0].Content, ShouldEqual, newArt.Content)
		So(narts[0].ViewCount, ShouldEqual, newMetas.ViewCount)
		So(narts[0].Title, ShouldEqual, newArt.Title)

		//删
		p3, err := d.EsDeleteArtMetas(ctx, art.Id)
		Convey("EsDeleteArtMetas err should be nil.p1 should not be nil.", func() {
			So(err, ShouldBeNil)
			So(p3, ShouldNotBeNil)
		})

		time.Sleep(time.Second * 2)
		darts, err := d.ArtMetasList(ctx, req)
		So(err, ShouldBeNil)
		So(len(darts), ShouldEqual, 0)

	})
}
