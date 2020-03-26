package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
)

func (d *Dao) GetMetasBySn(c context.Context, sn string) (res *model.Metas, err error) {
	res, err = d.GetCacheMetas(c, sn)
	addCache := true
	if err != nil {
		addCache = false
		err = nil
		log.SugarWithContext(c).Errorf("d.GetMetasBySn ERR:(%#+v),sn:(%#v),req(%#v)", err, sn, res)
	}
	defer func() {
		if res != nil && res.ArticleId == -1 {
			res = nil //注意这里只能修改命名返回值
		}
	}()
	//todo... add metrics
	if res != nil {
		return res, nil
	}
	res = new(model.Metas)
	cacheData := res
	cacheTime := 0
	//todo use distribute lock to protect db
	if err := d.db.Where("sn=?", sn).First(&res).Error; err != nil {
		cacheData = &model.Metas{ArticleId: -1, Sn: sn}
		cacheTime = utils.TimeHourSecond
	}
	if addCache {
		d.cacheQueue.Do(c, func(ctx context.Context) {
			d.SetCacheMetas(c, cacheData, cacheTime)
		})
	}
	return res, errors.Wrap(err, "d.GetMetasBySn")
}
