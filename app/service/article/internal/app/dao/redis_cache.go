package dao

import (
	"blog-go-microservice/app/service/article/internal/app/model/db"
	"context"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	etcdlock "github.com/zuiqiangqishao/framework/pkg/utils/lock/etcd"
	"strings"
)

const ART_PREFIX = "art_"
const META_PREFIX = "meta_"

const META_CHANGE_SET = "meta_change_set"

const TAG_ALL_KEY = "tag_all"

/***********************  文章缓存   *********************************/
func (d *Dao) GetCacheArticle(c context.Context, sn string) (*db.Article, error) {
	conn := d.redis.Get()
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("GET", ArtCacheKey(sn)))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrap(err, "GetCacheArticle GET err")
	}

	art := new(db.Article)
	if err = utils.JsonUnmarshal(reply, art); err != nil { //缓存没挂，log真实err，返回固定err让上层知道缓存没挂
		log.SugarWithContext(c).Errorf("d.GetCacheArticle sn:(%s),reply:(#%v),err:(%#+v)", sn, reply, errors.WithStack(err))
		return nil, ecode.JsonErr
	}

	return art, nil
}

//timeS is Number of seconds of cache duration, if it is zero the cache is forever
func (d *Dao) SetCacheArt(c context.Context, val *db.Article, timeS int) error {
	if val == nil {
		return nil
	}
	conn := d.redis.Get()
	defer conn.Close()
	json, err := utils.JsonMarshal(val)
	if err != nil {
		return errors.Wrap(err, "SetCacheArt jsonMarshal err:")
	}
	args := []interface{}{
		ArtCacheKey(val.Sn),
		json,
	}
	if timeS != 0 {
		args = append(args, "EX", timeS)
	}
	_, err = conn.Do("SET", args...)
	return errors.Wrap(err, "SetCacheArt SET Command err")
}

func (d *Dao) DeleteCacheArt(c context.Context, sn string) error {
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", ArtCacheKey(sn))
	return errors.Wrap(err, "DeleteCacheArt DEL err")
}

//************************  metas缓存  ***********************************************/
func (d *Dao) GetCacheMetas(c context.Context, sn string) (*db.Metas, error) {
	conn := d.redis.Get()
	defer conn.Close()
	reply, err := redis.Int64Map(conn.Do("HGETALL", MetasCacheKey(sn)))
	if err == redis.ErrNil || len(reply) == 0 {
		return nil, nil
	}
	res := &db.Metas{
		ArticleId: reply[db.ArtIdRedisKey],
		Sn:        sn,
		ViewCount: reply[db.ViewRedisKey],
		CmCount:   reply[db.CmRedisKey],
		LaudCount: reply[db.LaudRedisKey],
	}

	return res, errors.Wrap(err, "GetMetas HGETALL err")
}

func (d *Dao) SetCacheMetas(c context.Context, metas *db.Metas, timeS int) error {
	if metas == nil {
		return nil
	}
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", MetasCacheKey(metas.Sn),
		db.ArtIdRedisKey, metas.ArticleId,
		db.ViewRedisKey, metas.ViewCount,
		db.CmRedisKey, metas.CmCount,
		db.LaudRedisKey, metas.LaudCount,
	)
	if err != nil {
		return errors.Wrap(err, "SetMetas HMSET err")
	}
	if timeS != 0 {
		_, err = conn.Do("expire", MetasCacheKey(metas.Sn), timeS)
	}

	return errors.Wrap(err, "SetMetas Expire HMSET Key err")
}

func (d *Dao) IncCacheMetas(c context.Context, sn string, field string) error {
	conn := d.redis.Get()
	defer conn.Close()
	handler := func() error {
		_, err := conn.Do("HINCRBY", MetasCacheKey(sn), field, 1)
		if err != nil {
			return errors.Wrap(err, "AddMetas HINCRBY Err")
		}
		//metas缓存发生变化，加入到该Set,定时任务会定时同步该set里面的metas信息到ES、DB中
		_, err = conn.Do("SADD", MetasChangeSetKey(), sn)
		if err != nil {
			return errors.Wrap(err, "AddMetas HSET Err")
		}
		return nil
	}
	//分布式锁保证没有设置metas缓存的情况下，并发从DB中去取数据覆盖写缓存问题
	for i := 0; i < 5; i++ {
		//先看有没有
		e, err := redis.Int(conn.Do("exists", MetasCacheKey(sn)))
		if err != nil {
			return errors.Wrap(err, "AddMetas exisis Err")
		}
		if e == 0 {
			//没有就抢锁
			key := db.IncLockKey(sn)
			lockTTL := 5    //锁持续时间 s
			sleepTTL := 300 //锁争抢失败持续，睡眠时间
			m, err := etcdlock.DistributeTryLock(c, key, lockTTL, sleepTTL)

			if err != nil { //etcd server 挂了
				log.SugarWithContext(c).Errorf("etcdlock.DistributeTryLock ERR:(%#+v),sn:(%#v)", err, sn)
				break
			}

			if m == nil {
				continue
			}

			//lock success
			defer m.Release(c)
			break
		} else {
			return handler()
		}
	}
	metas, _ := d.GetMetasFromDB(c, "sn=?", sn)
	t := 0
	if metas == nil {
		metas = new(db.Metas)
		t = utils.TimeHourSecond
	}
	if err := d.SetCacheMetas(c, metas, t); err != nil {
		return err
	}
	return handler()
}

func (d *Dao) DelCacheMetas(c context.Context, sn string) error {
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", MetasCacheKey(sn))
	return errors.Wrap(err, "DelMetas DEL err")
}

/***************************  Tag缓存  ********************************************/
func (d *Dao) GetAllTagCache(c context.Context) ([]*db.Tag, error) {
	conn := d.redis.Get()
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", TAG_ALL_KEY))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrap(err, "GetAllTagCache GET err")
	}

	var res []*db.Tag
	if err = utils.JsonUnmarshal(reply, &res); err != nil {
		log.SugarWithContext(c).Errorf("d.GetAllTagCache sn:(%s),reply:(#%v),err:(%#+v)", reply, errors.WithStack(err))
		return nil, ecode.JsonErr
	}

	return res, nil
}

func (d *Dao) SetAllTagCache(c context.Context, tags []*db.Tag) error {
	conn := d.redis.Get()
	defer conn.Close()
	data, err := utils.JsonMarshal(tags)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = conn.Do("SET", TAG_ALL_KEY, data)
	return errors.WithStack(err)
}

func ArtCacheKey(sn string) string {
	return ART_PREFIX + sn
}
func MetasCacheKey(sn string) string {
	return META_PREFIX + sn
}

//metas缓存发生变化，加入到该Set,定时任务会定时同步该set里面的metas信息到ES、DB中
func MetasChangeSetKey() string {
	return META_CHANGE_SET
}

func PickSn(key string) (string, error) {
	s := strings.Split(key, "_")
	if len(s) < 2 {
		return "", errors.New("invalid key:" + key)
	}
	return s[1], nil
}

func ArtKeyWild() string {
	return ART_PREFIX + "*"
}
func MetasKeyWild() string {
	return META_PREFIX + "*"
}
