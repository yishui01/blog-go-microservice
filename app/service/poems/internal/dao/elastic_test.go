package dao

import (
	"blog-go-microservice/app/service/poems/internal/model"
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"testing"
)

func TestDaoCheckFilter(t *testing.T) {
	Convey("CheckFilter", t, func() {
		var (
			str = "abc,a,d,f|test,1,2,3"
		)
		p1, err := CheckFilter(str)
		So(err, ShouldNotBeNil)
		So(len(p1), ShouldEqual, 0)

		str = ",a,d,f"
		p1, err = CheckFilter(str)
		So(err, ShouldNotBeNil)
		So(len(p1), ShouldEqual, 0)

		str = "sn,1,2,3"
		p1, err = CheckFilter(str)
		So(err, ShouldBeNil)
		So(len(p1), ShouldEqual, 1)
		So(p1["sn"], ShouldResemble, []interface{}{"1", "2", "3"})

		str = "sn,1,2,3|content,5,7,6"
		p1, err = CheckFilter(str)
		So(err, ShouldBeNil)
		So(len(p1), ShouldEqual, 2)
		So(p1["sn"], ShouldResemble, []interface{}{"1", "2", "3"})
		So(p1["content"], ShouldResemble, []interface{}{"5", "7", "6"})

		str = "sn,1,2,3|content,5,,6"
		p1, err = CheckFilter(str)
		So(err, ShouldBeNil)
		So(len(p1), ShouldEqual, 2)
		So(p1["sn"], ShouldResemble, []interface{}{"1", "2", "3"})
		So(p1["content"], ShouldResemble, []interface{}{"5", "6"})

		str = "sn,1,2,3|,5,6"
		p1, err = CheckFilter(str)
		So(err, ShouldNotBeNil)
		So(len(p1), ShouldEqual, 1)
		So(p1["sn"], ShouldResemble, []interface{}{"1", "2", "3"})

	})
}

func TestDaoEsSearch(t *testing.T) {
	Convey("EsSearch", t, func() {
		var (
			c   = context.Background()
			req = model.Query{
				PageRequest: utils.PageRequest{PageSize: 10, PageNum: 1},
			}
		)

		r1 := req
		r1.Filter = ","
		poems, total, err := d.EsSearch(c, &r1)
		So(err.(*ecode.Status).Code(), ShouldEqual, ecode.RequestErr.Code())
		So(len(poems), ShouldEqual, 0)
		So(total, ShouldEqual, 0)

		r2 := req
		r2.Filter = "cate,shi-tang|author,李白"
		poems, total, err = d.EsSearch(c, &r2)
		So(err, ShouldBeNil)
		if total > 0 {
			for _, v := range poems {
				if v.Author != "李白" || v.Cate != "shi-tang" {
					t.Error("search res is not correct,res-", v, " filter-", r1.Filter)
				}
			}
		}

		if total >= 2 { //测试随机性
			oldId := poems[0].Id
			random := false
			for i := 0; i < 10; i++ {
				poems, total, err = d.EsSearch(c, &r2)
				if len(poems) > 0 && poems[0].Id != oldId {
					random = true
					break
				}
			}
			if !random {
				t.Error("res is not random")
			}
		}

	})
}
