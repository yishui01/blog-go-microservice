package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/utils"
)

const ART_PREFIX = "art_"
const TAG_PREFIX = "tag_"

const META_PREFIX = "meta_"

//文章缓存
func (d *Dao) GetCacheArticle(c context.Context, sn string) (*model.Article, error) {
	conn := d.redis.Get()
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("GET", ArtCacheKey(sn)))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrap(err, "GetCacheArticle GET err")
	}

	art := new(model.Article)
	if err = utils.JsonUnmarshal(reply, art); err != nil {
		return nil, errors.Wrap(err, "GetCacheArticle err")
	}

	return art, nil
}

//timeS is Number of seconds of cache duration, if it is zero the cache is forever
func (d *Dao) AddCacheArt(c context.Context, val *model.Article, timeS int) error {
	if val == nil {
		return nil
	}
	conn := d.redis.Get()
	defer conn.Close()
	json, err := utils.JsonMarshal(val)
	if err != nil {
		return errors.Wrap(err, "AddCacheArt jsonMarshal err:")
	}
	args := []interface{}{
		ArtCacheKey(val.Sn),
		json,
	}
	if timeS != 0 {
		args = append(args, "EX", timeS)
	}
	_, err = conn.Do("SET", args...)
	return errors.Wrap(err, "AddCacheArt SET Command err")
}

func (d *Dao) DeleteCacheArt(c context.Context, sn string) error {
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", ArtCacheKey(sn))
	return errors.Wrap(err, "DeleteCacheArt DEL err")
}

//metas缓存
func (d *Dao) GetCacheMetas(c context.Context, sn string) (*model.Metas, error) {
	conn := d.redis.Get()
	defer conn.Close()
	reply, err := redis.Int64Map(conn.Do("HGETALL", MetasCacheKey(sn)))
	if err == redis.ErrNil || len(reply) == 0 {
		return nil, nil
	}
	res := &model.Metas{
		ArticleId: reply[model.ArtIdRedisKey],
		Sn:        sn,
		ViewCount: reply[model.ViewRedisKey],
		CmCount:   reply[model.CmRedisKey],
		LaudCount: reply[model.LaudRedisKey],
	}

	return res, errors.Wrap(err, "GetMetas HGETALL err")
}

func (d *Dao) SetCacheMetas(c context.Context, metas *model.Metas, timeS int) error {
	if metas == nil {
		return nil
	}
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", MetasCacheKey(metas.Sn),
		model.ArtIdRedisKey, metas.ArticleId,
		model.ViewRedisKey, metas.ViewCount,
		model.CmRedisKey, metas.CmCount,
		model.LaudRedisKey, metas.LaudCount,
	)
	if err != nil {
		return errors.Wrap(err, "SetMetas HMSET err")
	}
	if timeS != 0 {
		_, err = conn.Do("expire", MetasCacheKey(metas.Sn), timeS)
	}

	return errors.Wrap(err, "SetMetas Expire HMSET Key err")
}

func (d *Dao) AddCacheMetas(c context.Context, sn string, field string) error {
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("HINCRBY", MetasCacheKey(sn), field, 1)
	return errors.Wrap(err, "AddMetas HINCRBY Err")
}

func (d *Dao) DelCacheMetas(c context.Context, sn string) error {
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", MetasCacheKey(sn))
	return errors.Wrap(err, "DelMetas DEL err")
}

func ArtCacheKey(sn string) string {
	return ART_PREFIX + sn
}
func MetasCacheKey(sn string) string {
	return META_PREFIX + sn
}
