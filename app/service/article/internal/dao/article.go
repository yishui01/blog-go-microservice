package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
)

type ListResp struct {
	Total int64
	Lists []*model.Article
	Page  int64
	Size  int64
}

func (d *Dao) ArtList(c context.Context, req *model.ArtQueryReq) (*ListResp, error) {
	res := new(ListResp)
	if req.KeyWords != "" {

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
	if err = d.db.Where("sn= ? ", sn).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound { //todo... add metrics
			d.cacheQueue.Do(c, func(ctx context.Context) { //数据库查不到就存一个，防止缓存穿透
				d.AddCacheArt(c, &model.Article{Id: -1, Sn: sn}, utils.TimeHourSecond)
			})
		}
		return nil, err
	}

	return res, nil
}

func (d *Dao) SaveArt(c context.Context, art *model.Article) (*model.Article, error) {
	var err error

	if err = d.db.Save(art).Error; err != nil {
		log.ZapLogger.Error("DB Add Art Err On SaveArt :" + err.Error())
		return nil, errors.WithStack(err)
	}
	if err = d.AddCacheArt(c, art, 0); err != nil {
		log.ZapLogger.Error("Redis Add ArtCache Err On SaveArt :" + err.Error())
		return nil, errors.WithStack(err)
	}
	//添加到es
	if _, err := d.PutArtToEs(c, art); err != nil {
		log.ZapLogger.Error("ES PUT Art Err:" + err.Error())
		return nil, errors.WithStack(err)
	}

	return art, nil
}
