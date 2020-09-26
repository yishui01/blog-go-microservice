package dao

import (
	"blog-go-microservice/app/service/article/internal/app/model/db"
	"flag"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/zuiqiangqishao/framework/pkg/setting"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"os"
	"strconv"
	"strings"
	"testing"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	flag.Set("conf", "../../test/")
	setting.Init()
	d, _ = New()
	InitCreateTags()
	code := m.Run()
	DeleteTags()
	os.Exit(code)
}

var (
	sourceTagIdStr   string
	sourceTagNameStr string
	sourceTags       []*db.Tag
	err              error
)

//先添加几个tag，测试的时候要用
func InitCreateTags() {
	if sourceTagIdStr, sourceTagNameStr, sourceTags, err = CreateTags(); err != nil {
		panic("create tags err" + err.Error())
	}
}

func DeleteTags() {
	d.db.Exec("DELETE FROM mc_tag where id in (?)", sourceTagIdStr)
}

func CreateTags() (tagIdStr, tagNameStr string, sourceTag []*db.Tag, err error) {
	tagIds := []string{}
	tagName := []string{}
	sourceTags := []*db.Tag{}
	num := 4
	for i := 0; i < num; i++ {
		name := "test-" + strconv.Itoa(i)
		t := db.Tag{Name: name}
		if err := d.db.Create(&t).Error; err != nil {
			return "", "", nil, err
		}
		tagIds = append(tagIds, strconv.FormatInt(t.Id, 10))
		tagName = append(tagName, name)
		sourceTags = append(sourceTags, &t)
	}
	return strings.Join(tagIds, ","), strings.Join(tagName, ","), sourceTags, nil
}

//
func TestCheckExist(t *testing.T) {
	Convey("UpdateArtMetas", t, func() {
		Convey("CheckExist", func() {
			art := &db.Article{Content: "999"}
			table := "mc_article"
			err := d.db.Table(table).Create(&art).Error
			So(err, ShouldBeNil)

			exists, err := utils.CheckExist(d.db, table, "id = ?", art.Id)
			Convey("Then err should be nil.id should not be nil.", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldEqual, true)
			})
			exists, err = utils.CheckExist(d.db, table, "id = -5")
			Convey("Then err should be nil.id should not be nil2.", func() {
				So(err, ShouldBeNil)
				So(exists, ShouldEqual, false)
			})
		})
	})
}
