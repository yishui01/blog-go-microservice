package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/gomodule/redigo/redis"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"go.uber.org/zap"
)

const ART_PREFIX = "art_"
const TAG_PREFIX = "tag_"

func (d *Dao) GetCacheArticle(c context.Context, sn string) (*model.Article, error) {
	conn := d.redis.Get()
	defer conn.Close()

	art := new(model.Article)
	reply, err := redis.Bytes(conn.Do("GET", ArtCacheKey(sn)))
	if err == nil {
		if err = utils.JsonUnmarshal(reply, art); err != nil {
			art = nil
		}
	}

	if err != nil && err != redis.ErrNil {
		log.ZapLogger.Error("GetCacheArticle err:", zap.String("err", err.Error()))
	}

	return art, err
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
		log.ZapLogger.Error("AddCacheArt jsonMarshal err:", zap.String("err", err.Error()))
		return err
	}
	args := []interface{}{
		ArtCacheKey(val.Sn),
		json,
	}
	if timeS != 0 {
		args = append(args, "EX", timeS)
	}
	_, err = conn.Do("SET", args...)
	if err != nil {
		log.ZapLogger.Error("AddCacheArt SET Command err:", zap.String("err", err.Error()))
	}
	return err
}

func (d *Dao) DeleteCacheArt(c context.Context, sn string) error {
	conn := d.redis.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", ArtCacheKey(sn))
	if err != nil {
		log.ZapLogger.Error("DeleteCacheArt err:", zap.String("err", err.Error()))
	}
	return err
}

func ArtCacheKey(sn string) string {
	return ART_PREFIX + sn
}
