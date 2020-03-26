package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	etcdlock "github.com/zuiqiangqishao/framework/pkg/utils/lock/etcd"
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
			err := utils.JsonUnmarshal(hit.Source, &t)
			if err != nil {
				log.ZapWithContext(c).Error("ES Select Art json.Unmarshal resp Err On ArtList :"+err.Error(), zap.String("hit", utils.Vf(hit)))
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
		log.SugarWithContext(c).Errorf("d.GetCacheArticle ERR:(%#+v),sn:(%#v),req(%#v)", err, sn, res)
		err = nil
		addCache = false //cache挂了就不要往里加缓存了
	}
	defer func() {
		if res != nil && res.Id == -1 {
			res = nil //注意这里只能修改命名返回值
		}
	}()
	if res != nil { //todo... add cache hit metrics
		return res, nil
	}

	//todo... add cache lose metrics
	res = new(model.Article)
	cacheData := res
	cacheTime := 0 //forever

	if addCache {
		//没有缓存，并且redis没挂，etcd分布式锁 防止缓存击穿/穿透
		key := model.ArtLockKey(sn)
		lockTTL := 5
		retry := 3
		failLockSleepMill := 500
		for i := 0; i < retry && addCache; i++ {
			cacheRes, err := d.GetCacheArticle(c, sn)
			if cacheRes != nil {
				return cacheRes, nil
			}
			if err != nil { //cache挂了
				log.SugarWithContext(c).Errorf("d.GetCacheArticle ERR on DistLock:(%#+v),sn:(%#v),req(%#v)", err, sn, res)
				addCache = false
				break
			}
			//try lock is nonblock
			m, err := etcdlock.DistributeTryLock(c, key, lockTTL, failLockSleepMill)
			if err != nil { //etcd server 挂了
				log.SugarWithContext(c).Errorf("etcdlock.DistributeTryLock ERR:(%#+v),sn:(%#v),req(%#v)", err, sn, res)
				break
			}

			if m != nil { //lock success
				defer m.Release(c)
				break
			}
		}
	}

	//到这里来有以下几种情况
	//1、lock success，
	//2、已到获取锁最大次数，不再尝试获取
	//3、cache 挂了
	//4、etcd server 挂了
	if err = d.db.Where("sn= ? ", sn).First(&res).Error; err != nil {
		cacheData = &model.Article{Id: -1, Sn: sn}
		cacheTime = utils.TimeHourSecond

		//todo... db err add metrics
		if err == gorm.ErrRecordNotFound {
			err = ecode.NothingFound
		} else {
			err = fmt.Errorf("DB Select Art Err On GetArtBySn  sn=(%s),err=(%v)", sn, err)
		}
	}

	if addCache {
		d.cacheQueue.Do(c, func(ctx context.Context) {
			d.AddCacheArt(c, cacheData, cacheTime)
		})
	}

	return res, errors.WithStack(err)
}

//DB创建文章+metas
func (d *Dao) CreateArt(c context.Context, art *model.Article, metas *model.Metas) (id int64, err error) {
	db := d.db
	art.Id = 0 //Omit ID Column
	tx := db.Begin()
	if err = tx.Create(&art).Error; err != nil {
		tx.Rollback()
		return art.Id, errors.WithStack(err)
	}
	metas.ArticleId = art.Id
	metas.Sn = art.Sn
	if err = tx.Create(metas).Error; err != nil {
		tx.Rollback()
		return art.Id, errors.WithStack(err)
	}
	tx.Commit()

	return art.Id, nil
}

func (d *Dao) CheckArtExist(art *model.Article) (bool, error) {
	res := new(model.Article)
	if err := d.db.Table("mc_article").Where("id=?", art.Id).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, errors.WithStack(err)
	}

	return true, nil
}

//DB修改文章+metas
func (d *Dao) UpdateArt(c context.Context, art *model.Article, metas *model.Metas) (id int64, err error) {
	b, err := d.CheckArtExist(art)
	if err != nil {
		return art.Id, err
	}
	if !b {
		return art.Id, ecode.NothingFound
	}
	tx := d.db.Begin()
	if err = tx.Table("mc_metas").Where("article_id=?", metas.ArticleId).Update(map[string]interface{}{
		"view_count": metas.ViewCount,
		"cm_count":   metas.CmCount,
		"laud_count": metas.LaudCount,
	}).Error; err != nil {
		tx.Rollback()
		return art.Id, errors.WithStack(err)
	}
	ArtMaps := map[string]interface{}{
		"title":   art.Title,
		"tags":    art.Tags,
		"img":     art.Img,
		"content": art.Content,
		"status":  art.Status,
	}
	if art.CreatedAt.Second() > 0 {
		ArtMaps["created_at"] = art.CreatedAt
	}
	if art.UpdatedAt.Second() > 0 {
		ArtMaps["updated_at"] = art.UpdatedAt
	}

	if err = tx.Table("mc_article").Where("id=?", art.Id).Update(ArtMaps).Error; err != nil {
		tx.Rollback()
		return art.Id, errors.WithStack(err)
	}
	tx.Commit()
	return art.Id, nil
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
	if err = tx.Where("id = ?", id).First(art).Error; err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}
	if err = tx.Where("id = ?", id).Delete(&model.Article{}).Error; err != nil && err != gorm.ErrRecordNotFound {
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
