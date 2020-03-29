package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/garyburd/redigo/redis"
	"github.com/olivere/elastic/v7"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/sync/errgroup"
	"strconv"
	"sync"
	"time"
)

//后台定时任务
var _cron = Cron{}

var _defaultJobs = []CommJob{
	{Name: "metasSync", Time: time.Second, Run: (*Dao).MetasSync},
}

type CommJob struct {
	Name      string
	Time      time.Duration
	Run       func(d *Dao)
	closeChan chan struct{}
}

type Cron struct {
	once   sync.Once
	jobs   []*CommJob
	closed bool
	chanWg sync.WaitGroup
}

func (d *Dao) CronStart(c context.Context) (close func()) {
	_cron.once.Do(func() {
		g := errgroup.WithContext(c)
		for _, v := range _defaultJobs {
			g.Go(func(ctx context.Context) error {
				t := time.NewTicker(v.Time)
				for {
					if _cron.closed {
						return nil
					}
					select {
					case <-t.C:
						if err := d.jobQueue.Do(c, func(c context.Context) {
							v.Run(d)
						}); err != nil {
							log.SugarWithContext(nil).Warnf("d.jobQueue.Doerr:(%#+v)", err)
						}
					}
				}
			})
		}
	})
	return _cron.Close
}

func (c *Cron) Close() {
	c.closed = true
	c.chanWg.Wait()
}

//同步metas信息到数据库、ES
func (d *Dao) MetasSync() {
	conn := d.redis.Get()
	defer conn.Close()
	keys, err := redis.Strings(conn.Do("keys", MetasKeyWild()))
	if err != nil {
		log.SugarWithContext(context.TODO()).Warnf("Cron d.MetasSync err keys(%v),err:(%v)", keys, err)
	}
	var metas *model.Metas
	bulkRequest := d.es.Bulk().Index(model.ART_ES_INDEX).Type("_doc")
	for _, v := range keys {
		sn, err := PickSn(v)
		if err != nil || sn == "" {
			log.SugarWithContext(context.TODO()).Warnf("Cron d.MetasSync PickSn err:(%#+v),v(%s)", err, v)
			continue
		}
		metas, err = d.GetCacheMetas(context.TODO(), sn)

		if metas == nil {
			return
		}

		//DB
		if err := d.db.Table("mc_metas").Where("article_id=?", metas.ArticleId).Update(map[string]int64{
			"view_count": metas.ViewCount,
			"cm_count":   metas.CmCount,
			"laud_count": metas.LaudCount,
		}).Error; err != nil {
			log.SugarWithContext(nil).Errorf("d.MetasSync Err(%#v)", err)
		}

		bulkRequest.Add(elastic.NewBulkUpdateRequest().Doc(map[string]int64{
			"view_count": metas.ViewCount,
			"cm_count":   metas.CmCount,
			"laud_count": metas.LaudCount,
		}).
			Id(strconv.FormatInt(metas.ArticleId, 10)))
	}

	//ES
	bulkResp, err := bulkRequest.Do(context.TODO())
	if err != nil {
		log.SugarWithContext(nil).Warnf("Cron MetasSync:err=(%#v),bulkResp=(%#v)", err, bulkResp)
	}
}
