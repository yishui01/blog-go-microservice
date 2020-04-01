package dao

import (
	"blog-go-microservice/app/service/webinfo/internal/model"
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strconv"
	"testing"
)

func TestDaoGetInfoListDB(t *testing.T) {
	Convey("CreateInfoDB", t, func() {
		var (
			c  = context.Background()
			ss = []*model.WebInfo{
				{WebKey: "abc", UniqueVal: "9", WebVal: "1", Status: 1, Ord: "10"},
				{WebKey: "abc", UniqueVal: "2", WebVal: "1", Status: 2, Ord: "5"},
				{WebKey: "abc", UniqueVal: "3", WebVal: "1", Status: 3, Ord: "5"},
				{WebKey: "abc", UniqueVal: "4", WebVal: "1", Status: 5, Ord: "5"},
				{WebKey: "abc", UniqueVal: "5", WebVal: "1", Status: 5, Ord: "6"},
				{WebKey: "abc", UniqueVal: "6", WebVal: "1", Status: 5, Ord: "7"},
			}

			query = Query{
				PageRequest: utils.PageRequest{PageSize: int32(len(ss)), PageNum: 1},
			}
			info = &model.WebInfo{WebKey: "abc", UniqueVal: "999", WebVal: "1", Status: 1, Ord: "10"}
			p1   int64
		)
		Convey("When everything goes positive", func() {
			err := d.db.Exec("DELETE FROM mc_web_info").Error //先清空数据表
			So(err, ShouldBeNil)

			//增
			p1, err = d.CreateInfoDB(c, info)
			So(err, ShouldBeNil)
			So(p1, ShouldBeGreaterThan, 0)
			p1, err = d.CreateInfoDB(c, info)
			serr := err.(*ecode.Status)
			So(serr.Code(), ShouldEqual, ecode.UniqueErr.Code())

			for _, v := range ss {
				p1, err := d.CreateInfoDB(c, v)
				So(err, ShouldBeNil)
				So(p1, ShouldBeGreaterThan, 0)
			}

			//查
			res, err := d.GetInfoListDB(context.TODO(), &query) //基本分页查询
			fmt.Println(res[0], res[1], res[2])
			So(err, ShouldBeNil)
			So(len(res), ShouldEqual, query.PageSize)
			So(res[0].Id > res[1].Id, ShouldBeTrue)

			t := query
			t.Filter = "id," + strconv.FormatInt(res[0].Id, 10) //测试过滤
			res, err = d.GetInfoListDB(context.TODO(), &t)
			So(err, ShouldBeNil)
			So(len(res), ShouldEqual, 1)

			t = query
			t.PageNum = 1
			t.PageSize = int32(5)
			t.Order = "status|desc,id|asc" //测试排序
			res, err = d.GetInfoListDB(context.TODO(), &t)
			So(err, ShouldBeNil)
			So(len(res), ShouldEqual, t.PageSize)
			So(res[0].Status, ShouldEqual, 5)
			So(res[0].Ord, ShouldEqual, "5")
			So(res[1].Status, ShouldEqual, 5)
			So(res[1].Ord, ShouldEqual, "6")
			So(res[2].Status, ShouldEqual, 5)
			So(res[2].Ord, ShouldEqual, "7")

			//改
			bak := ss[0]
			bak.WebKey = ss[1].WebKey
			bak.UniqueVal = ss[1].UniqueVal
			id, err := d.UpdateInfoDB(c, bak)
			serr = err.(*ecode.Status)
			So(serr.Code(), ShouldEqual, ecode.UniqueErr.Code())

			bak.WebVal = "5"
			bak.WebKey = "6"
			id, err = d.UpdateInfoDB(c, bak)
			So(id, ShouldBeGreaterThan, 0)
			So(err, ShouldBeNil)
			t = query
			t.Filter = "id," + strconv.FormatInt(bak.Id, 10)
			infos, err := d.GetInfoListDB(context.TODO(), &t) //基本分页查询
			So(err, ShouldBeNil)
			So(len(infos), ShouldEqual, 1)
			So(infos[0].WebKey, ShouldEqual, bak.WebKey)
			So(infos[0].WebVal, ShouldEqual, bak.WebVal)
			So(infos[0].UniqueVal, ShouldEqual, bak.UniqueVal)

			//删除
			err = d.DeleteInfoDB(c, bak.Id, false)
			So(err, ShouldBeNil)
			query.Unscoped = true
			t = query
			t.Filter = "id," + strconv.FormatInt(bak.Id, 10)
			infos, err = d.GetInfoListDB(c, &t)
			So(err, ShouldBeNil)
			So(len(infos), ShouldEqual, 1)
			So(infos[0].DeletedAt.Unix(), ShouldBeGreaterThan, 0)

			err = d.DeleteInfoDB(c, bak.Id, true)
			query.Unscoped = true
			infos, err = d.GetInfoListDB(c, &t)
			So(err, ShouldBeNil)
			So(len(infos), ShouldEqual, 0)

		})

	})

}

func TestDaoBuildFilter(t *testing.T) {
	Convey("BuildFilter", t, func() {
		var (
			c         = context.Background()
			filterStr = ""
		)
		p1, err := d.BuildFilter(c, filterStr, d.db)
		Convey("Then err should be nil.p1 should not be nil.", func() {
			So(err, ShouldBeNil)
			So(p1, ShouldNotBeNil)
		})
	})
}

func TestDaoBuildOrder(t *testing.T) {
	Convey("BuildOrder", t, func() {
		var (
			c        = context.Background()
			orderStr = ""
		)
		p1, err := d.BuildOrder(c, orderStr, d.db)
		Convey("Then err should be nil.p1 should not be nil.", func() {
			So(err, ShouldBeNil)
			So(p1, ShouldNotBeNil)
		})
	})
}
