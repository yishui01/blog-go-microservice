package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoGetTagAll(t *testing.T) {
	Convey("tag operation", t, func() {
		var (
			c = context.Background()
		)

		tag := &model.Tag{Name: strconv.FormatInt(time.Now().Unix(), 10) + "create55"}
		copyTag := &model.Tag{Name: tag.Name}

		//*********************创建Tag**************************/
		p2, err := d.CreateTag(c, tag)
		So(err, ShouldBeNil)
		So(p2, ShouldBeGreaterThan, 0)
		p3, err := d.CreateTag(c, copyTag)
		So(errors.Cause(err), ShouldEqual, ecode.UniqueErr)
		So(p3, ShouldEqual, -1)

		//获取Tag
		getAllTags, err := d.GetTagAll(c)
		So(err, ShouldBeNil)
		flag := false
		for _, v := range getAllTags {
			if v.Id == tag.Id {
				flag = true
				break
			}
		}
		if flag == false {
			t.Error("get tag cache err after createtag", err)
		}

		//测试GetFirstTagFromDB
		p1, err := d.GetFirstTagFromDB(c, map[string]interface{}{"id": tag.Id})
		Convey("GetFirstTagFromDB err should be nil.p1 should not be nil.", func() {
			So(err, ShouldBeNil)
			So(p1.Id, ShouldBeGreaterThan, 0)
		})

		//测试 GetAllTagFromDB
		p1s, err := d.GetAllTagFromDB(c, map[string]interface{}{"id": tag.Id})
		Convey("GetAllTagFromDB err should be nil.p1 should not be nil.", func() {
			So(err, ShouldBeNil)
			So(len(p1s), ShouldBeGreaterThan, 0)
		})

		//加入文章中间关联
		art := &model.Article{Content: "tagdelete", Tags: strconv.FormatInt(tag.Id, 10)}
		artId, err := d.CreateArtMetas(c, art, nil)
		So(err, ShouldBeNil)
		So(artId, ShouldBeGreaterThan, 0)

		//***********************修改Tag*******************************/
		newTag := &model.Tag{Id: tag.Id, Name: strconv.FormatInt(time.Now().Unix(), 10) + "update88"}
		tagId, err := d.UpdateTag(c, newTag)
		So(err, ShouldBeNil)
		So(tagId, ShouldBeGreaterThan, 0)

		getAllTags, err = d.GetTagAll(c)
		So(err, ShouldBeNil)
		flag = false
		for _, v := range getAllTags {
			if v.Id == tagId && v.Name == newTag.Name {
				flag = true
				break
			}
		}
		if flag == false {
			t.Error("get tag cache err after updatetag", err)
		}

		newArt, err := d.GetArtFromDB(c, "id=?", art.Id) //db art的tag字段
		So(err, ShouldBeNil)
		So(newArt.Tags, ShouldEqual, newTag.Name)

		resArt, err := d.GetCacheArticle(c, newArt.Sn) //cache art的tag字段
		So(err, ShouldBeNil)
		So(resArt.Tags, ShouldEqual, newTag.Name)

		artTag, err := d.GetArtTagsFromDB(c, art.Id) //中间表Name字段
		So(err, ShouldBeNil)
		So(artTag[0].TagName, ShouldEqual, newTag.Name)

		Convey("UpdateTag err should be nil.tagId should not be nil.", func() {
			So(err, ShouldBeNil)
			So(tagId, ShouldBeGreaterThan, 0)
		})

		//***************************************软删除Tag********************/
		err = d.DeleteTag(c, -1, false)
		So(errors.Cause(err), ShouldEqual, ecode.RequestErr)

		err = d.DeleteTag(c, tagId, false)
		So(err, ShouldBeNil)

		//缓存里没有了
		getAllTags, err = d.GetTagAll(c)
		So(err, ShouldBeNil)
		flag = false
		for _, v := range getAllTags {
			if v.Id == tag.Id {
				flag = true
				break
			}
		}
		if flag {
			t.Error("get tag cache err after DeleteTag", err)
		}

		//查看tag表
		_, err = d.GetFirstTagFromDB(c, map[string]interface{}{"id": tag.Id})
		So(errors.Cause(err), ShouldEqual, gorm.ErrRecordNotFound)
		unt := new(model.Tag)
		err = d.db.Unscoped().Table("mc_tag").Where("id=?", tag.Id).First(unt).Error
		So(err, ShouldBeNil)
		So(unt.Id, ShouldEqual, tag.Id) //这样可以查到

		//查看中间表的tag关联
		dArtTag := new(model.ArticleTag)
		aerr := d.db.Table("mc_article_tag").Where("tag_id=? and article_id=?", tag.Id, art.Id).
			First(dArtTag).Error
		if err != nil {
			So(errors.Cause(aerr), ShouldEqual, gorm.ErrRecordNotFound)
		}

		//查看关联的文章的tag字段
		newArt, err = d.GetArtFromDB(c, "id=?", art.Id) //db art的tag字段
		So(err, ShouldBeNil)
		So(newArt.Tags, ShouldEqual, "")

		//***********************硬删除Tag******************************/
		err = d.DeleteTag(c, tagId, true)
		So(err, ShouldBeNil)
		//记录已经物理删除
		err = d.db.Unscoped().Table("mc_tag").Where("id=?", tagId).First(unt).Error
		So(errors.Cause(err), ShouldEqual, gorm.ErrRecordNotFound)
	})
}

func TestDaoRefreshRelateArt(t *testing.T) {
	Convey("RefreshRelateArt", t, func() {
		var (
			c     = context.Background()
			tagId = []int64{0}
		)
		Convey("When everything goes positive", func() {
			err := d.RefreshRelateArt(c, tagId)
			Convey("Then err should be nil.", func() {
				So(errors.Cause(err), ShouldEqual, gorm.ErrRecordNotFound)
			})
		})
	})
}

func TestDaoRefreshTagAllCache(t *testing.T) {
	Convey("RefreshTagAllCache", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			err := d.RefreshTagAllCache(c)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
