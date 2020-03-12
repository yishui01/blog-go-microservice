package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
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

func (d *Dao) ArtList(c context.Context, req *model.ArtQueryReq) ([]*model.EsArticle, error) {
	res := make([]*model.EsArticle, 0)
	esResp, err := d.EsSearchArt(c, req)
	if err != nil {
		log.ZapLogger.Error("ES Select Art Err On ArtList :" + err.Error())
		return res, err
	}
	if esResp.TotalHits() > 0 {
		for _, hit := range esResp.Hits.Hits {
			var t model.EsArticle
			err := json.Unmarshal(hit.Source, &t)
			if err != nil {
				log.ZapLogger.Error("ES Select Art json.Unmarshal resp Err On ArtList :"+err.Error(), zap.String("doc_id", hit.Id))
				continue
			}
			res = append(res, &t)
		}
	}

	return res, nil
}

func (d *Dao) GetArtBySn(c context.Context, sn string) (*model.Article, error) {
	res, err := d.GetCacheArticle(c, sn)
	if err != nil {
		err = nil
	}
	defer func() {
		if res != nil && res.Id == -1 {
			res = nil
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
		if err != gorm.ErrRecordNotFound {
			//todo... db err add metrics
			log.ZapLogger.Error("DB Select Art Err On GetArtBySn :" + err.Error())
			err = errors.WithStack(err)
		}
	}

	d.cacheQueue.Do(c, func(ctx context.Context) {
		d.AddCacheArt(c, cacheData, cacheTime)
	})
	return res, err
}

func (d *Dao) SaveArt(c context.Context, art *model.Article) (*model.Article, error) {
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
	if err = db.Save(art).Error; err != nil {
		log.ZapLogger.Error("DB Add Art Err On SaveArt :" + err.Error())
		return nil, errors.WithStack(err)
	}
	//把最新的数据查出来，不然可能没有created_at 和updated_at
	if err := db.Where("id = ?", art.Id).First(art).Error; err != nil {
		log.ZapLogger.Error("DB Select new insert Art Err On SaveArt :" + err.Error())
		return nil, errors.WithStack(err)
	}
	if err = d.AddCacheArt(c, art, 0); err != nil {
		log.ZapLogger.Error("Redis Add ArtCache Err On SaveArt :" + err.Error())
		return nil, errors.WithStack(err)
	}
	if _, err := d.EsPutArt(c, art); err != nil {
		log.ZapLogger.Error("ES PUT Art Err:" + err.Error())
		return nil, errors.WithStack(err)
	}

	return art, nil
}

func (d *Dao) DelArt(c context.Context, id int64, physical bool) error {
	db := d.db
	if physical {
		db = d.db.Unscoped()
	}
	art := new(model.Article)
	if err := db.Where("id = ?", id).First(art).Error; err != nil {
		log.ZapLogger.Error("DB select art on delete DelArt Err:" + err.Error())
		return errors.WithStack(err)
	}
	if err := db.Where("id = ?", id).Delete(&model.Article{}).Error; err != nil {
		log.ZapLogger.Error("DB delete DelArt Err:" + err.Error())
		return errors.WithStack(err)
	}
	if err := d.DeleteCacheArt(c, art.Sn); err != nil {
		log.ZapLogger.Error("redis delete DelArt DeleteCacheArt Err:" + err.Error())
		return errors.WithStack(err)
	}
	if physical {
		if _, err := d.EsDeleteArt(c, art.Id); err != nil {
			log.ZapLogger.Error("ES Physical delete DelArt  Err:" + err.Error())
			return errors.WithStack(err)
		}
	} else {
		art.DeletedAt = time.Now()
		if _, err := d.EsPutArt(c, art); err != nil {
			log.ZapLogger.Error("ES Soft delete DelArt  Err:" + err.Error())
			return errors.WithStack(err)
		}
	}

	return nil
}
