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
	"strconv"
	"strings"
	"time"
)

type ListResp struct {
	Total int64
	Lists []*model.Article
	Page  int64
	Size  int64
}

//只从ES中查找文章列表
func (d *Dao) ArtMetasList(c context.Context, req *model.ArtQueryReq) ([]*model.EsArticle, error) {
	res := make([]*model.EsArticle, 0)
	esResp, err := d.EsSearchArtMetas(c, req)
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

//获取文章详情,不包含metas
func (d *Dao) GetArtBySn(c context.Context, sn string) (res *model.Article, err error) {
	res, err = d.GetCacheArticle(c, sn)
	addCache := true
	if err != nil {
		if !ecode.EqualError(ecode.JsonErr, err) {
			log.SugarWithContext(c).Errorf("d.GetCacheArticle ERR:(%#+v),sn:(%#v),req(%#v)", err, sn, res)
			addCache = false //cache挂了就不要往里加缓存了
		}
		err = nil
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
			if err != nil {
				if !ecode.EqualError(ecode.JsonErr, err) {
					//cache挂了
					log.SugarWithContext(c).Errorf("d.GetCacheArticle ERR on DistLock:(%#+v),sn:(%#v),req(%#v)", err, sn, res)
					addCache = false
					break
				}
				err = nil
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
	res, err = d.GetArtFromDB(c, "sn=?", sn)
	if err != nil {
		cacheData = &model.Article{Id: -1, Sn: sn}
		cacheTime = utils.TimeHourSecond
	}
	if addCache {
		if err := d.cacheQueue.Do(c, func(c context.Context) {
			if err := d.SetCacheArt(c, cacheData, cacheTime); err != nil {
				log.SugarWithContext(c).Errorf("d.SetCacheArt Err(%#+v)", err)
			}
		}); err != nil {
			log.SugarWithContext(c).Errorf("d.cacheQueue.Do Err(%#+v)", err)
		}
	}

	return res, errors.WithStack(err)
}

//DB创建文章+metas+tags中间表维护
func (d *Dao) CreateArtMetas(c context.Context, art *model.Article, metas *model.Metas) (artId int64, err error) {
	var (
		db         = d.db
		tagNameStr = ""
		tags       = []*model.Tag{}
	)
	art.Id = 0 //Omit ID Column
	if metas == nil {
		metas = new(model.Metas)
	}

	//1、找出tag
	if len(art.Tags) > 0 {
		if tags, tagNameStr, err = d.extractTagFromIdStr(art.Tags); err != nil {
			return art.Id, err
		}
		art.Tags = tagNameStr
	}

	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			if perr := recover(); perr != nil {
				tx.Rollback()
				panic(perr) //这里panic应该往上层抛出
			}
			err = errors.WithStack(tx.Commit().Error)
		}
	}()

	//2、添加文章
	if err = tx.Create(&art).Error; err != nil {
		return art.Id, errors.WithStack(err)
	}
	metas.ArticleId = art.Id
	metas.Sn = art.Sn
	//3、添加metas
	if err = tx.Create(metas).Error; err != nil {
		return art.Id, errors.WithStack(err)
	}
	//4、维护article_tag中间表
	if len(tagNameStr) > 0 {
		if err = d.updateRelationArtTag(art.Id, tags, tx); err != nil {
			return art.Id, err
		}
	}

	return art.Id, nil
}

func (d *Dao) extractTagFromIdStr(tagIdStr string) ([]*model.Tag, string, error) {
	tagIds := []int64{}
	res := []*model.Tag{}
	tagNameSlice := []string{}
	uniqueMaps := make(map[int]bool)
	for _, v := range strings.Split(tagIdStr, ",") {
		tagId, err := strconv.Atoi(v)
		if err != nil {
			return nil, "", errors.Wrap(ecode.RequestErr, "tag错误："+err.Error())
		}
		if uniqueMaps[tagId] {
			continue
		}
		uniqueMaps[tagId] = true
		tagIds = append(tagIds, int64(tagId))
	}
	for _, tagId := range tagIds {
		t := &model.Tag{}
		if err := d.db.Table("mc_tag").Where("id = ?", tagId).First(&t).Error; err != nil {
			return nil, "", errors.Wrap(ecode.RequestErr, "tag错误："+err.Error())
		}
		res = append(res, t)
		tagNameSlice = append(tagNameSlice, t.Name)
	}
	return res, strings.Join(tagNameSlice, ","), nil
}

func (d *Dao) updateRelationArtTag(artId int64, tags []*model.Tag, tx *gorm.DB) error {
	if tx == nil {
		return errors.New("tx can not be nil")
	}
	var err error
	//带tx参数的函数不要再tx.Begin()开事务了，因为这个是被调用者，强行开会 can't start transaction

	//更新中间表，先delete再insert，由于是更新关联，接Exec硬删除即可
	if err = tx.Exec("DELETE FROM mc_article_tag Where article_id=?", artId).Error; err != nil {
		return errors.WithStack(err)
	}

	for _, tag := range tags {
		if err = tx.Table("mc_article_tag").Create(&model.ArticleTag{ArticleId: artId,
			TagId: tag.Id, TagName: tag.Name}).Error; err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

//DB修改文章+metas+tags
func (d *Dao) UpdateArtMetas(c context.Context, art *model.Article, metas *model.Metas) (id int64, err error) {
	b, err := utils.CheckExist(d.db, "mc_article", "id = ?", art.Id)
	if err != nil {
		return art.Id, err
	}
	if !b {
		return art.Id, ecode.NothingFound
	}

	tx := d.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			if perr := recover(); perr != nil {
				tx.Rollback()
				panic(perr)
			}
			err = errors.WithStack(tx.Commit().Error)
		}
	}()

	if err = tx.Table("mc_metas").Where("article_id=?", metas.ArticleId).Update(map[string]interface{}{
		"view_count": metas.ViewCount,
		"cm_count":   metas.CmCount,
		"laud_count": metas.LaudCount,
	}).Error; err != nil {
		return art.Id, errors.WithStack(err)
	}

	var tagNameStr = ""

	if len(art.Tags) > 0 {
		var (
			tags = []*model.Tag{}
			err  error
		)
		if tags, tagNameStr, err = d.extractTagFromIdStr(art.Tags); err != nil {
			return art.Id, err
		}
		if err := d.updateRelationArtTag(art.Id, tags, tx); err != nil {
			return art.Id, err
		}
	}

	ArtMaps := map[string]interface{}{
		"title":   art.Title,
		"tags":    tagNameStr,
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
		return art.Id, errors.WithStack(err)
	}
	return art.Id, nil
}

//删除文章
func (d *Dao) DelArtMetas(c context.Context, id int64, physical bool) (err error) {
	if id <= 0 {
		return errors.New("id is invalid")
	}
	db := d.db
	if physical {
		db = d.db.Unscoped()
	}
	tx := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			if perr := recover(); perr != nil {
				tx.Rollback()
				panic(perr) //这里panic应该往上层抛出
			}
			err = errors.WithStack(tx.Commit().Error)
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

	if err = d.DelCacheMetas(c, art.Sn); err != nil {
		return err
	}

	if physical {
		if _, err = d.EsDeleteArtMetas(c, art.Id); err != nil {
			return err
		}
	} else {
		art.DeletedAt = time.Now()
		if _, err = d.EsUpdateArtMetas(c, art, nil); err != nil {
			return err
		}
	}
	return nil
}

//刷新文章tags冗余字段、文章缓存，metas缓存、ES数据
func (d *Dao) RefreshArt(c context.Context, artId int64) error {
	art := new(model.Article)
	if err := d.db.Where("id = ?", artId).First(art).Error; err != nil {
		return errors.WithStack(err)
	}
	tags, err := d.GetArtTagsFromDB(c, artId)
	if err != nil {
		return err
	}
	art.Tags = model.BuildArtTagStr(tags)
	if err := d.db.Table("mc_article").Where("id=?", art.Id).Update(map[string]interface{}{
		"tags": art.Tags,
	}).Error; err != nil {
		return err
	}

	if err := d.SetCacheArt(c, art, 0); err != nil {
		return err
	}
	metas, err := d.GetMetasFromDB(c, "sn = ?", art.Sn)
	if err != nil {
		return err
	}
	if _, err := d.EsPutArtMetas(c, art, metas); err != nil {
		return err
	}
	return nil
}

func (d *Dao) GetArtTagsFromDB(c context.Context, artId int64) ([]*model.ArticleTag, error) {
	res := make([]*model.ArticleTag, 0)
	err := d.db.Table("mc_article_tag").Where("article_id =?", artId).Find(&res).Error
	return res, errors.WithStack(err)
}

func (d *Dao) GetArtFromDB(c context.Context, query interface{}, args ...interface{}) (res *model.Article, err error) {
	res = new(model.Article)
	if err = d.db.Where(query, args...).First(&res).Error; err != nil {
		res = nil
		//todo... db err add metrics
		if err == gorm.ErrRecordNotFound {
			err = ecode.NothingFound
		} else {
			err = fmt.Errorf("DB Select Art Err On GetArtFromDB  query=(%#v), args=(%#v),err=(%v)", query, args, err)
		}
	}
	err = errors.WithStack(err)
	return
}

func (d *Dao) GetMetasFromDB(c context.Context, query interface{}, args ...interface{}) (res *model.Metas, err error) {
	res = new(model.Metas)
	if err = d.db.Where(query, args...).First(&res).Error; err != nil {
		//todo... db err add metrics
		if err == gorm.ErrRecordNotFound {
			err = ecode.NothingFound
		} else {
			err = fmt.Errorf("DB Select Metas Err On GetMetasFromDB  query=(%#v),args=(%#v),err=(%v)", query, args, err)
		}
	}
	err = errors.WithStack(err)
	return
}

//异步累加metas
func (d *Dao) AddMetasCount(c context.Context, sn string, field string) error {
	return d.cacheQueue.Do(c, func(c context.Context) {
		if err := d.IncCacheMetas(c, sn, field); err != nil {
			log.SugarWithContext(c).Warnf("AddMetasCount.IncCacheMetas Err(%#v)", err)
		}
	})
}
