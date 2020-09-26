package dao

import (
	"blog-go-microservice/app/service/article/internal/app/model/db"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
)

func (d *Dao) GetMetasBySn(c context.Context, sn string) (res *db.Metas, err error) {
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
	res = new(db.Metas)
	cacheData := res
	cacheTime := 0
	//todo use distribute lock to protect db
	if err = d.db.Where("sn=?", sn).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = ecode.NothingFound
		}
		res = nil
		cacheData = &db.Metas{ArticleId: -1, Sn: sn}
		cacheTime = utils.TimeHourSecond
	}
	if addCache { //里面的err
		if cerr := d.cacheQueue.Do(c, func(c context.Context) {
			if err := d.SetCacheMetas(c, cacheData, cacheTime); err != nil {
				log.SugarWithContext(c).Error("d.cacheQueue.Do d.SetCacheMetas Err(%#v)", err)
			}
		}); cerr != nil {
			log.SugarWithContext(c).Error("d.cacheQueue.Do Err(%#v)", cerr)
		}
	}
	return res, errors.Wrap(err, "d.GetMetasBySn")
}
