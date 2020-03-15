package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"go.uber.org/zap"
	"time"
)

type ListResp struct {
	Total int64
	Lists []*model.Article
	Page  int64
	Size  int64
}

//只从ES中查找文章列表
func (d *Dao) ArtList(c context.Context, req *model.ArtQueryReq) ([]*model.EsArticle, error) {
	res := make([]*model.EsArticle, 0)
	esResp, err := d.EsSearchArt(c, req)
	if err != nil {
		return res, err
	}
	if esResp.TotalHits() > 0 {
		for _, hit := range esResp.Hits.Hits {
			var t model.EsArticle
			err := json.Unmarshal(hit.Source, &t)
			if err != nil {
				log.ZapLogger.Error("ES Select Art json.Unmarshal resp Err On ArtList :"+err.Error(), zap.String("hit", utils.Vf(hit)))
				continue
			}
			res = append(res, &t)
		}
	}

	return res, nil
}

//获取文章详情
func (d *Dao) GetArtBySn(c context.Context, sn string) (res *model.Article, err error) {
	res, err = d.GetCacheArticle(c, sn)
	addCache := true
	if err != nil {
		log.SugarLogger.Errorf("d.GetCacheArticle ERR:(%#+v),sn:(%#v),req(%#v)", err, sn, res)
		err = nil
		addCache = false //cache挂了就不要往里加缓存了
	}
	defer func() {
		if res != nil && res.Id == -1 {
			res = nil //注意这里只能修改命名返回值
		}
	}()
	//todo... add metrics
	if res != nil {
		return res, nil
	}

	res = new(model.Article)
	cacheData := res
	cacheTime := 0 //forever
	if err = d.db.Where("sn= ? ", sn).First(&res).Error; err != nil {
		cacheData = &model.Article{Id: -1, Sn: sn} //防止缓存击穿
		cacheTime = utils.TimeHourSecond

		//todo... db err add metrics
		if err == gorm.ErrRecordNotFound {
			err = ecode.NothingFound
		} else {
			err = errors.Wrap(err, "DB Select Art Err On GetArtBySn  sn="+sn)
		}
	}

	if addCache {
		d.cacheQueue.Do(c, func(ctx context.Context) {
			d.AddCacheArt(c, cacheData, cacheTime)
		})
	}

	return res, err
}

//保存文章到DB,未设置缓存
func (d *Dao) SaveArt(c context.Context, art *model.Article, metas *model.Metas) (*model.Article, error) {
	var err error
	db := d.db
	if art.Id > 0 {
		db = db.Omit("sn")
	}
	if art.CreatedAt.IsZero() {
		db = db.Omit("created_at")
	}
	if art.UpdatedAt.IsZero() {
		db = db.Omit("updated_at")
	}

	tx := db.Begin()
	if err = db.Save(art).Error; err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	metas.ArticleId = art.Id
	metas.Sn = art.Sn
	if err = db.Save(metas).Error; err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	tx.Commit()
	return art, nil
}

//刷新文章缓存，ES数据
func (d *Dao) SetArtCache(c context.Context, artId int64) error {
	art := new(model.Article)
	db := d.db
	if err := db.Where("id = ?", artId).First(art).Error; err != nil {
		return errors.WithStack(err)
	}
	if err := d.AddCacheArt(c, art, 0); err != nil {
		return err
	}
	if _, err := d.EsPutArt(c, art); err != nil {
		return err
	}
	return nil
}

//删除文章
func (d *Dao) DelArt(c context.Context, id int64, physical bool) error {
	db := d.db
	if physical {
		db = d.db.Unscoped()
	}
	var err error
	tx := db.Begin()
	defer func() {
		if pe := recover(); pe != nil || err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	art := new(model.Article)
	if err = db.Where("id = ?", id).First(art).Error; err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}
	if err = db.Where("id = ?", id).Delete(&model.Article{}).Error; err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}

	if err = d.DeleteCacheArt(c, art.Sn); err != nil {
		return err
	}
	if physical {
		if _, err = d.EsDeleteArt(c, art.Id); err != nil {
			return err
		}
	} else {
		art.DeletedAt = time.Now()
		if _, err = d.EsPutArt(c, art); err != nil {
			return err
		}
	}
	return nil
}
