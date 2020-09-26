package article

import (
	"blog-go-microservice/app/service/article/internal/app/dao"
	"blog-go-microservice/app/service/article/internal/app/model/db"
	pb "blog-go-microservice/app/service/article/pb"
	"bou.ke/monkey"
	"context"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"reflect"
	"testing"
)

func TestServiceTagList(t *testing.T) {

	Convey("TagList", t, func() {
		var (
			c   = context.Background()
			req = &pb.TagListReq{}
		)
		Convey("When everything goes positive", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetTagAll", func(_ *dao.Dao, _ context.Context) ([]*db.Tag, error) {
				return []*db.Tag{{Id: 1}}, nil
			})
			defer guard.Unpatch()
			p1, err := s.TagList(c, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				So(len(p1.Lists), ShouldEqual, 1)
				So(p1.Lists[0].Id, ShouldEqual, 1)
			})
		})
		Convey("When everything goes positive2", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "GetTagAll", func(_ *dao.Dao, _ context.Context) ([]*db.Tag, error) {
				return nil, errors.New("aaa")
			})
			defer guard.Unpatch()
			p1, err := s.TagList(c, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldNotBeNil)
				So(p1, ShouldBeNil)
			})
		})
	})
}

func TestServiceCreateTag(t *testing.T) {
	Convey("CreateTag", t, func() {
		var (
			c   = context.Background()
			req = &pb.SaveTagReq{}
		)
		Convey("When everything goes positive", func() {

			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshTagAllCache", func(_ *dao.Dao, c context.Context) error {
				return nil
			})
			defer guard.Unpatch()

			Convey("Then err should be nil.p1 should not be nil.", func() {
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
					return 1, nil
				})
				defer guard.Unpatch()
				p1, err := s.CreateTag(c, req)
				So(err, ShouldBeNil)
				So(p1.Data, ShouldEqual, "1")
			})

			Convey("Then err should be nil.p1 should not be nil2.", func() {
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
					return 1, errors.New("111")
				})
				defer guard.Unpatch()
				p1, err := s.CreateTag(c, req)
				So(errors.Cause(err).(*ecode.Status).Code(), ShouldEqual, ecode.ServerErr.Code())
				So(p1, ShouldBeNil)

				guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
					return 1, ecode.RequestErr
				})
				p1, err = s.CreateTag(c, req)
				So(err, ShouldEqual, ecode.RequestErr)
				So(p1, ShouldBeNil)
			})
		})
	})
}

func TestServiceUpdateTag(t *testing.T) {
	Convey("UpdateTag", t, func() {
		var (
			c   = context.Background()
			req = &pb.SaveTagReq{}
		)

		Convey("When everything goes positive", func() {

			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshTagAllCache", func(_ *dao.Dao, c context.Context) error {
				return nil
			})
			defer guard.Unpatch()

			Convey("Then err should be nil.p1 should not be nil.", func() {
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "UpdateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
					return 1, nil
				})
				defer guard.Unpatch()
				p1, err := s.UpdateTag(c, req)
				So(err, ShouldBeNil)
				So(p1.Data, ShouldEqual, "1")
			})

			Convey("Then err should be nil.p1 should not be nil2.", func() {
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "UpdateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
					return 1, errors.New("111")
				})
				defer guard.Unpatch()
				p1, err := s.UpdateTag(c, req)
				So(errors.Cause(err).(*ecode.Status).Code(), ShouldEqual, ecode.ServerErr.Code())
				So(p1, ShouldBeNil)

				guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "UpdateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
					return 1, ecode.RequestErr
				})
				p1, err = s.UpdateTag(c, req)
				So(err, ShouldEqual, ecode.RequestErr)
				So(p1, ShouldBeNil)
			})

			Convey("Then err should be nil.p1 should not be nil3.", func() {
				guard = monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "RefreshTagAllCache", func(_ *dao.Dao, c context.Context) error {
					return errors.New("RefreshTagAllCache err")
				})
				guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "UpdateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
					return 1, nil
				})
				defer guard.Unpatch()
				p1, err := s.UpdateTag(c, req)
				So(err, ShouldBeNil)
				So(p1.Data, ShouldEqual, "1")
				So(p1.Status, ShouldEqual, 1)
			})
		})

	})
}

func TestServicesaveTag(t *testing.T) {
	Convey("saveTag", t, func() {
		var (
			c        = context.Background()
			req      = &pb.SaveTagReq{}
			isUpdate bool
		)
		Convey("When everything goes positive", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "CreateTag", func(_ *dao.Dao, c context.Context, tag *db.Tag) (int64, error) {
				return 1, errors.New("111")
			})
			defer guard.Unpatch()
			p1, err := s.saveTag(c, req, isUpdate)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err.(ecode.Codes).Code(), ShouldEqual, ecode.ServerErr.Code())
				So(p1, ShouldBeNil)
			})
		})
	})
}

func TestServiceDeleteTag(t *testing.T) {
	Convey("DeleteTag", t, func() {
		var (
			c   = context.Background()
			req = &pb.DelTagReq{}
		)
		Convey("When everything goes positive", func() {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(s.dao), "DeleteTag", func(_ *dao.Dao, c context.Context, id int64, physical bool) (err error) {
				return errors.New("111")
			})
			defer guard.Unpatch()
			p1, err := s.DeleteTag(c, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldNotBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

//
func TestServiceAssignTag(t *testing.T) {
	Convey("AssignTag", t, func() {
		var (
			req = &pb.SaveTagReq{}
		)
		Convey("When everything goes positive", func() {
			p1 := AssignTag(req)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
