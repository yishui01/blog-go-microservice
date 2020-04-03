package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDaoextractTagFromIdStr(t *testing.T) {
	Convey("extractTagFromIdStr", t, func() {
		var (
			InvalidTag  = "1a"
			NotFoundTag = "1,-1"
		)
		Convey("When everything goes positive", func() {
			_, _, err := d.extractTagFromIdStr(InvalidTag)
			Convey("invalid tag test", func() {
				So(errors.Cause(err), ShouldEqual, ecode.RequestErr)
			})
			_, _, err = d.extractTagFromIdStr(NotFoundTag)
			Convey("NotFoundTag tag test", func() {
				So(errors.Cause(err), ShouldEqual, ecode.RequestErr)
			})

			tags, tagNameStr, err := d.extractTagFromIdStr(sourceTagIdStr)

			Convey("Then err should be nil.p1,p2 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(tagNameStr, ShouldEqual, sourceTagNameStr)
				So(sourceTagIdStr, ShouldNotBeNil)
				for k, v := range sourceTags {
					if tags[k].Id != v.Id || tags[k].Name != v.Name {
						t.Errorf("ERR:sourceTags(%#v),tag(%#v)", sourceTags, tags)
						t.Errorf("ERR:*tags[k] (%#v),*v(%#v)", *tags[k], *v)
					}
				}
			})

		})
	})
}

func TestDaoupdateRelationArtTag(t *testing.T) {
	Convey("updateRelationArtTag", t, func() {
		var (
			tags = []*model.Tag{
				&model.Tag{Name: "tname1", Id: 1},
				&model.Tag{Name: "tname2", Id: 2},
				&model.Tag{Name: "tname3", Id: 3},
			}
		)
		Convey("When everything goes positive", func() {
			art := model.Article{Content: "111"}
			if d.db.Create(&art).Error != nil {
				t.Error("create art err", err)
			}

			err := d.updateRelationArtTag(art.Id, tags, nil)
			Convey("Then err should not be nil.", func() {
				So(err, ShouldNotBeNil)
			})

			tx := d.db.Begin()
			err = d.updateRelationArtTag(0, tags, tx)
			tx.Rollback()
			Convey("Then artId is 0 cause err should not be nil.", func() {
				So(err, ShouldNotBeNil)
			})

			tx = d.db.Begin()
			err = d.updateRelationArtTag(art.Id, tags, tx)
			tx.Commit()

			Convey("Then err should not be nil and tag is equal.", func() {
				So(err, ShouldBeNil)
				res := make([]*model.ArticleTag, 0)
				d.db.Table("mc_article_tag").Where("article_id = ?", art.Id).Find(&res)
				for k, _ := range res {
					if res[k].TagId != tags[k].Id || res[k].TagName != tags[k].Name {
						t.Errorf("res[k](%#v),tags[k](%#v)", res[k], tags[k])
					}
				}
			})

			d.db.Exec("DELETE FROM mc_article_tag where article_id = ?", art.Id)
			d.db.Exec("DELETE FROM mc_article where id=?", art.Id)
		})
	})
}

func TestDaoCreateArtMetas(t *testing.T) {
	Convey("CreateArtMetas", t, func() {
		//先创建两个tag
		tags := []*model.Tag{}
		tagIds := []string{}
		tagNames := []string{}
		key := "testaaa"
		for i := 0; i < 2; i++ {
			name := strconv.Itoa(i) + key
			a := &model.Tag{Name: name}
			d.db.Table("mc_tag").Create(&a)
			tags = append(tags, a)
			if a.Id == 0 {
				t.Errorf("add tag err tag(%#v)", a)
			}
			tagIds = append(tagIds, strconv.FormatInt(a.Id, 10))
			tagNames = append(tagNames, name)
		}
		tagIdStr := strings.Join(tagIds, ",")
		tagNameStr := strings.Join(tagNames, ",")
		Convey("When everything goes positive", func() {
			var (
				c     = context.Background()
				art   = &model.Article{Id: 1, Sn: "123456", Content: "666666", Tags: tagIdStr}
				metas = &model.Metas{ViewCount: 9, LaudCount: 19, CmCount: 99}
			)
			artId, err := d.CreateArtMetas(c, art, metas)
			Convey("Then err should be nil.artId should not be nil.", func() {
				So(err, ShouldBeNil)
				So(artId, ShouldBeGreaterThan, 0)

				var (
					content string
					tags    string
					laud    int64
					cname   string
					dname   string
				)
				rows, err := d.db.Raw(
					"SELECT a.content,a.tags,b.laud_count,c.tag_name as cname, d.name as dname "+
						" FROM mc_article as a "+
						" LEFT JOIN mc_metas as b on a.id = b.article_id"+
						" LEFT JOIN mc_article_tag as c on a.id = c.article_id"+
						" LEFT JOIN mc_tag as d on c.tag_id = d.id"+
						" WHERE a.id=?", art.Id).Rows()
				if err != nil {
					t.Errorf("select err(%#v)", err)
				}

				i := 0
				for rows.Next() {
					if err := rows.Scan(&content, &tags, &laud, &cname, &dname); err != nil {
						t.Errorf("select Err (%#v)", err)
					}
					if content != art.Content || art.Tags != tagNameStr || laud != metas.LaudCount {
						t.Error("equal Err", content, art.Content, art.Tags, tagNameStr, laud, metas.LaudCount)
					}
					if cname != dname || cname != strconv.Itoa(i)+key {
						t.Error("equal tag name Err", cname, dname, strconv.Itoa(i)+key)
					}
					i++
				}
			})

			d.db.Exec("DELETE FROM mc_article where id=?", art.Id)
			d.db.Exec("DELETE FROM mc_tag where id in (?)", tagIdStr)
			d.db.Exec("DELETE FROM mc_metas where article_id=?", art.Id)
			d.db.Exec("DELETE FROM mc_article_tag where article_id=?", art.Id)
		})
	})
}

func TestDaoArtMetasList(t *testing.T) {
	Convey("ArtMetasList", t, func() {
		var (
			c    = context.Background()
			req  = &model.ArtQueryReq{}
			req2 = &model.ArtQueryReq{
				PageRequest: utils.PageRequest{PageSize: 1, PageNum: 1},
				KeyWords:    "111",
				Tags:        "666",
				Status:      1,
				Order:       "view_count|desc",
				CreatedAt:   1,
				UpdatedAt:   1,
				Unscoped:    true,
			}
		)
		Convey("When everything goes positive", func() {
			p1, total, err := d.ArtMetasList(c, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				So(total, ShouldBeGreaterThan, -1)
			})
			p1, total, err = d.ArtMetasList(c, req2)
			Convey("Then err should be nil.p1 should not be nil2.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				So(total, ShouldBeGreaterThan, -1)
			})
		})
	})
}

func TestDaoGetArtBySn(t *testing.T) {
	Convey("GetArtBySn", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			sn := ""
			err := d.DeleteCacheArt(context.TODO(), sn)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
			res, err := d.GetArtBySn(c, sn)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err.Error(), ShouldEqual, ecode.NothingFound.Error())
				So(res, ShouldBeNil)
			})
			res, err = d.GetArtBySn(c, sn)
			Convey("Then err should be nil.res should be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldBeNil)
			})
		})

	})

	//测试加cache的异步任务
	Eventually(func() bool {
		res := false
		if err := d.cacheQueue.Do(context.TODO(), func(c context.Context) {
			if err := d.SetCacheArt(c, &model.Article{Id: -5, Sn: ""}, 30); err == nil {
				res = true
			}
		}); err != nil {
			t.Error("async create err", err)
		}
		art, err := d.GetCacheArticle(context.TODO(), "")
		return art != nil && art.Id == -5 && err == nil && res
	}).Should(BeTrue())
}

func TestDaoUpdateArtMetas(t *testing.T) {
	Convey("UpdateArtMetas", t, func() {
		var (
			c     = context.Background()
			art   = &model.Article{Content: "666777"}
			metas = &model.Metas{LaudCount: 9, CmCount: 7, ViewCount: 88}
		)
		Convey("When everything goes positive", func() {
			artId, err := d.CreateArtMetas(context.TODO(), art, metas)
			Convey("err is nil", func() {
				So(err, ShouldBeNil)
				So(artId, ShouldBeGreaterThan, 0)
				So(artId, ShouldBeGreaterThan, 0)
			})
			DBArt, err := d.GetArtFromDB(context.TODO(), "id=?", art.Id)
			DBMetas, err := d.GetMetasFromDB(context.TODO(), "article_id=?", art.Id)
			Convey("GetArtFromDB err is nil", func() {
				So(err, ShouldBeNil)
				So(DBArt.Id, ShouldBeGreaterThan, 0)
				So(DBArt.Content, ShouldEqual, art.Content)
				So(DBMetas.LaudCount, ShouldEqual, metas.LaudCount)
				So(DBArt.Tags, ShouldEqual, "")
			})

			art.Content = "8778988"
			art.Tags = sourceTagIdStr
			metas.ViewCount = 50

			id, err1 := d.UpdateArtMetas(c, art, metas)
			DBArt, err2 := d.GetArtFromDB(context.TODO(), "id=?", art.Id)
			DBMetas, err3 := d.GetMetasFromDB(context.TODO(), "article_id=?", art.Id)
			artTags, err4 := d.GetArtTagsFromDB(context.TODO(), art.Id)

			//看下article、metas，tag，tag中间表、article的tag冗余字段是否更新
			Convey("Then err should be nil.id should not be nil.", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(err3, ShouldBeNil)
				So(err4, ShouldBeNil)
				So(id, ShouldEqual, DBArt.Id)
				So(DBArt.Content, ShouldEqual, DBArt.Content)
				So(DBArt.Tags, ShouldEqual, sourceTagNameStr)
				So(DBMetas.ViewCount, ShouldEqual, metas.ViewCount)

				So(len(artTags), ShouldEqual, len(sourceTags))
				for k, _ := range artTags {
					So(artTags[k].TagName, ShouldEqual, sourceTags[k].Name)
					So(artTags[k].TagId, ShouldEqual, sourceTags[k].Id)
					So(artTags[k].ArticleId, ShouldEqual, art.Id)
				}
			})
		})
	})
}

func TestDaoDelArtMetas(t *testing.T) {
	Convey("DelArtMetas", t, func() {
		var (
			c = context.Background()
		)
		err := d.DelArtMetas(c, 0, false)
		Convey("Then err should not be nil.", func() {
			So(err, ShouldNotBeNil)
		})

		Convey("When everything goes positive", func() {
			art := model.Article{Title: "111"}
			metas := model.Metas{ViewCount: 111}
			if _, err := d.CreateArtMetas(context.TODO(), &art, &metas); err != nil {
				t.Error("create art err", err)
			}

			if err = d.RefreshArt(context.TODO(), art.Id); err != nil {
				t.Error("RefreshArt art err", err)
			}

			err = d.DelArtMetas(c, art.Id, false)
			f := func() ([]*model.EsArticle, int64, error) {
				return d.ArtMetasList(context.TODO(), &model.ArtQueryReq{
					PageRequest: utils.PageRequest{PageNum: 1, PageSize: 10},
					Terms:       "id," + strconv.FormatInt(art.Id, 10),
					Status:      0,
					Unscoped:    true,
				})
			}
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				t := new(model.Article)
				err := d.db.Unscoped().Where("id=?", art.Id).First(&t).Error
				So(err, ShouldBeNil)
				So(t.Title, ShouldEqual, art.Title)

				time.Sleep(time.Millisecond * 1000) //es索引可能不能立即生效，等一秒

				res, total, err := f()

				So(err, ShouldBeNil)
				So(len(res), ShouldEqual, 1)
				So(total, ShouldEqual, 1)
				So((*res[0]).Title, ShouldEqual, art.Title)
				So((*res[0]).DeletedAt.Second(), ShouldBeGreaterThan, 1)
			})
			err = d.DelArtMetas(c, art.Id, true)
			Convey("physical true Then err should be nil.", func() {
				So(err, ShouldBeNil)
				t := new(model.Article)
				err := d.db.Unscoped().Where("id=?", art.Id).First(&t).Error
				So(err, ShouldEqual, gorm.ErrRecordNotFound)

				time.Sleep(time.Millisecond * 500)
				res, total, err := f()
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 0)
				So(len(res), ShouldEqual, 0)

			})
		})
	})
}

func TestDaoRefreshArt(t *testing.T) {
	Convey("RefreshArt++GetArtTagsFromDB", t, func() {
		var (
			c     = context.Background()
			art   = &model.Article{Content: "777", Tags: sourceTagIdStr}
			metas = &model.Metas{ViewCount: 50}
		)

		artId, err := d.CreateArtMetas(context.TODO(), art, metas)
		if err != nil {
			t.Error("createArtMetas err", err)
		}

		//刷新ES、缓存、中间表关联
		Convey("RefreshArt", func() {
			err := d.RefreshArt(c, artId)
			Convey("RefreshArt err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})

		//获取art-tag中间表
		Convey("GetArtTagsFromDB", func() {
			p1, err := d.GetArtTagsFromDB(c, artId)
			Convey("GetArtTagsFromDB err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
				So(len(p1), ShouldEqual, len(sourceTags))
			})
		})

		res, err := d.GetArtFromDB(c, "id=?", artId)
		Convey("GetArtFromDB err should be nil.res should not be nil.", func() {
			So(err, ShouldBeNil)
			So(res.Id, ShouldEqual, art.Id)
			So(res.Content, ShouldEqual, art.Content)
			So(res.Tags, ShouldEqual, art.Tags)
		})

		num := 500
		for i := 0; i < num; i++ {
			if err = d.AddMetasCount(c, res.Sn, model.ViewRedisKey); err != nil {
				t.Error(err)
			}
		}

		//测试加cache的异步任务
		time.Sleep(time.Second * 5)
		d.MetasSync()
		Convey("AddMetasCount", func() {
			So(err, ShouldBeNil)
			m, err := d.GetMetasBySn(c, res.Sn)
			So(err, ShouldBeNil)
			So(m.ViewCount, ShouldEqual, metas.ViewCount+int64(num))
		})
		Convey("GetMetasFromDB  err should be nil.res should not be nil.", func() {
			d.MetasSync()
			res, err := d.GetMetasFromDB(c, "article_id=?", artId)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.ViewCount, ShouldEqual, metas.ViewCount+int64(num))
		})

	})
}
