package article

import (
	pb "blog-go-microservice/app/service/article/pb"
	"blog-go-microservice/app/service/article/internal/app/dao"
	"blog-go-microservice/app/service/article/internal/app/model/db"
	"bou.ke/monkey"
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"reflect"
	"testing"
)

func TestServiceArtList(t *testing.T) {
	Convey("ArtList", t, func() {
		var (
			ctx     = context.Background()
			listReq = &pb.ArtListRequest{}
		)
		Convey("When everything goes positive", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "ArtMetasList", func(_ *dao.Dao, _ context.Context, _ *db.ArtQueryReq) ([]*db.EsArticle, int64, error) {
				return nil, 0, nil
			})
			defer guard.Unpatch()
			p1, err := s.ArtList(ctx, listReq)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})

			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "ArtMetasList", func(_ *dao.Dao, _ context.Context, _ *db.ArtQueryReq) ([]*db.EsArticle, int64, error) {
				return []*db.EsArticle{
					{Id: 111},
				}, 0, nil
			})
			p1, err = s.ArtList(ctx, listReq)
			Convey("Then err should be nil.p1 should not be nil2.", func() {
				So(err, ShouldBeNil)
				So(len(p1.Lists), ShouldEqual, 1)
				So(p1.Lists[0].Id, ShouldEqual, 111)
			})
		})

		Convey("When everything goes positive2", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "ArtMetasList", func(_ *dao.Dao, _ context.Context, _ *db.ArtQueryReq) ([]*db.EsArticle, int64, error) {
				return nil, 0, ecode.RequestErr
			})
			defer guard.Unpatch()
			p1, err := s.ArtList(ctx, listReq)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldResemble, ecode.RequestErr)
				So(p1, ShouldBeNil)
			})
		})

		Convey("When everything goes positive3", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "ArtMetasList", func(_ *dao.Dao, _ context.Context, _ *db.ArtQueryReq) ([]*db.EsArticle, int64, error) {
				return nil, 0, errors.New("aaa")
			})
			defer guard.Unpatch()
			p1, err := s.ArtList(ctx, listReq)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldResemble, ecode.ServerErr)
				So(p1, ShouldBeNil)
			})
		})
	})
}

//
func TestServiceGetArtBySn(t *testing.T) {
	Convey("GetArtBySn", t, func() {
		var (
			ctx    = context.Background()
			artReq = &pb.ArtDetailRequest{}
		)
		Convey("When everything goes positive", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetArtBySn", func(_ *dao.Dao, _ context.Context, _ string) (_ *db.Article, _ error) {
				return nil, nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "AddMetasCount", func(_ *dao.Dao, c context.Context, sn string, field string) error {
				return nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetMetasBySn", func(_ *dao.Dao, c context.Context, sn string) (res *db.Metas, err error) {
				return nil, nil
			})
			defer guard.Unpatch()
			p1, err := s.GetArtBySn(ctx, artReq)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldResemble, ecode.NothingFound)
				So(p1, ShouldBeNil)
			})
		})

		Convey("When everything goes positive2", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetArtBySn", func(_ *dao.Dao, _ context.Context, _ string) (_ *db.Article, _ error) {
				return nil, ecode.RequestErr
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "AddMetasCount", func(_ *dao.Dao, c context.Context, sn string, field string) error {
				return nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetMetasBySn", func(_ *dao.Dao, c context.Context, sn string) (res *db.Metas, err error) {
				return nil, nil
			})
			defer guard.Unpatch()
			p1, err := s.GetArtBySn(ctx, artReq)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldResemble, ecode.ServerErr)
				So(p1, ShouldBeNil)
			})
		})

		Convey("When everything goes positive3", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetArtBySn", func(_ *dao.Dao, _ context.Context, _ string) (_ *db.Article, _ error) {
				return nil, ecode.NothingFound
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "AddMetasCount", func(_ *dao.Dao, c context.Context, sn string, field string) error {
				return nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetMetasBySn", func(_ *dao.Dao, c context.Context, sn string) (res *db.Metas, err error) {
				return nil, nil
			})
			defer guard.Unpatch()
			p1, err := s.GetArtBySn(ctx, artReq)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldResemble, ecode.NothingFound)
				So(p1, ShouldBeNil)
			})
		})

		Convey("When everything goes positive4", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetArtBySn", func(_ *dao.Dao, _ context.Context, _ string) (_ *db.Article, _ error) {
				return new(db.Article), nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "AddMetasCount", func(_ *dao.Dao, c context.Context, sn string, field string) error {
				return nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetMetasBySn", func(_ *dao.Dao, c context.Context, sn string) (res *db.Metas, err error) {
				return nil, nil
			})
			defer guard.Unpatch()
			p1, err := s.GetArtBySn(ctx, artReq)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				So(p1.Id, ShouldEqual, 0)
				So(p1.ViewCount, ShouldEqual, 0)
			})
		})

	})
}

//
func TestServiceAssignDetail(t *testing.T) {
	Convey("AssignDetail", t, func() {
		var (
			art   = &db.Article{}
			metas = &db.Metas{}
		)
		Convey("When everything goes positive", func() {
			p1 := AssignDetail(art, metas)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestServiceAssignSave(t *testing.T) {
	Convey("AssignSave", t, func() {
		var (
			req = &pb.SaveArtReq{}
		)
		Convey("When everything goes positive", func() {
			p1, p2 := AssignSave(req)
			Convey("Then p1,p2 should not be nil.", func() {
				So(p2, ShouldNotBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestServiceCreateArt(t *testing.T) {
	Convey("CreateArt", t, func() {
		var (
			ctx = context.Background()
			req = &pb.SaveArtReq{}
		)
		Convey("testCorrect", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateArtMetas", func(_ *dao.Dao, c context.Context, art *db.Article, metas *db.Metas) (artId int64, err error) {
				return 1, nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshArt", func(_ *dao.Dao, c context.Context, artId int64) error {
				return nil
			})
			defer guard.Unpatch()

			p1, err := s.CreateArt(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1.Data, ShouldEqual, "1")
				So(p1.Status, ShouldEqual, 0)
			})
		})

		Convey("test CreateArtMetas err", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshArt", func(_ *dao.Dao, c context.Context, artId int64) error {
				return nil
			})
			defer guard.Unpatch()

			Convey("testRequestErr", func() {
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateArtMetas", func(_ *dao.Dao, c context.Context, art *db.Article, metas *db.Metas) (artId int64, err error) {
					return 1, ecode.RequestErr
				})
				defer guard.Unpatch()
				p1, err := s.CreateArt(ctx, req)
				Convey("Then err should be nil.p1 should not be nil.", func() {
					So(err, ShouldEqual, ecode.RequestErr)
					So(p1, ShouldBeNil)
				})
			})

			Convey("testServerErr", func() {
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateArtMetas", func(_ *dao.Dao, c context.Context, art *db.Article, metas *db.Metas) (artId int64, err error) {
					return 1, errors.New("123")
				})
				defer guard.Unpatch()
				p1, err := s.CreateArt(ctx, req)
				Convey("Then err should be nil.p1 should not be nil.", func() {
					So(err.(*ecode.Status).Code(), ShouldEqual, ecode.ServerErr.Code())
					So(p1, ShouldBeNil)
				})
			})
		})

		Convey("test RefreshArt err", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateArtMetas", func(_ *dao.Dao, c context.Context, art *db.Article, metas *db.Metas) (artId int64, err error) {
				return 1, nil
			})

			defer guard.Unpatch()

			Convey("testRefreshArtErr", func() {
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshArt", func(_ *dao.Dao, c context.Context, artId int64) error {
					return errors.New("AAA")
				})
				defer guard.Unpatch()
				p1, err := s.CreateArt(ctx, req)
				Convey("Then err should be nil.p1 should not be nil.", func() {
					So(err, ShouldBeNil)
					So(p1.Status, ShouldEqual, 1)
				})
			})
		})
	})
}

func TestServiceUpdateArt(t *testing.T) {
	Convey("UpdateArt", t, func() {
		var (
			ctx = context.Background()
			req = &pb.SaveArtReq{}
		)

		Convey("testCorrect", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "UpdateArtMetas", func(_ *dao.Dao, c context.Context, art *db.Article, metas *db.Metas) (artId int64, err error) {
				return 1, nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshArt", func(_ *dao.Dao, c context.Context, artId int64) error {
				return nil
			})
			defer guard.Unpatch()

			p1, err := s.UpdateArt(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1.Data, ShouldEqual, "1")
				So(p1.Status, ShouldEqual, 0)
			})
		})

	})
}

func TestServiceSaveArt(t *testing.T) {
	Convey("SaveArt", t, func() {
		var (
			ctx      = context.Background()
			req      = &pb.SaveArtReq{}
			isUpdate bool
		)
		Convey("testCorrect", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateArtMetas", func(_ *dao.Dao, c context.Context, art *db.Article, metas *db.Metas) (artId int64, err error) {
				return 1, nil
			})
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshArt", func(_ *dao.Dao, c context.Context, artId int64) error {
				return nil
			})
			defer guard.Unpatch()

			p1, err := s.SaveArt(ctx, req, isUpdate)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1.Data, ShouldEqual, "1")
				So(p1.Status, ShouldEqual, 0)
			})
		})

	})
}

func TestServiceDeleteArt(t *testing.T) {
	Convey("DeleteArt", t, func() {
		var (
			ctx = context.Background()
			req = &pb.DelArtRequest{}
		)
		Convey("testCorrect", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "DelArtMetas", func(_ *dao.Dao, c context.Context, id int64, physical bool) (err error) {
				return nil
			})
			defer guard.Unpatch()

			p1, err := s.DeleteArt(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})

		Convey("testDeleteErr", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "DelArtMetas", func(_ *dao.Dao, c context.Context, id int64, physical bool) (err error) {
				return errors.New("aa")
			})
			defer guard.Unpatch()

			p1, err := s.DeleteArt(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldNotBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestServiceClose(t *testing.T) {
	Convey("Close", t, func() {
		Convey("When everything goes positive", func() {
			s.Close()
			Convey("No return values", func() {
			})
		})
	})
}

func TestServicePing(t *testing.T) {
	Convey("Ping", t, func() {
		var (
			ctx = context.Background()
			e   = &empty.Empty{}
		)
		Convey("When everything goes positive", func() {
			p1, err := s.Ping(ctx, e)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
