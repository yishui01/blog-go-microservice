package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
)

//查询Tag，一次性取出
func (d *Dao) GetTagAll(c context.Context) ([]*model.Tag, error) {
	res, err := d.GetAllTagCache(c)
	addCache := true
	if err != nil {
		if !ecode.EqualError(ecode.JsonErr, err) {
			log.SugarWithContext(c).Errorf("d.TagAll Err:(%#+v)", err)
			addCache = false
		}
		err = nil
	}
	if res != nil { //todo... add cache hit metrics
		return res, nil
	}

	res = []*model.Tag{}
	err = d.db.Table("mc_tags").Find(&res).Error
	if err != nil {
		return res, errors.WithStack(err)
	}
	if addCache {
		err = d.cacheQueue.Do(c, func(c context.Context) {
			if err := d.SetAllTagCache(c, res); err != nil {
				log.SugarWithContext(c).Errorf("d.SetAllTagCache Err:(%#+v)", err)
			}
		})
	}

	return res, errors.WithStack(err)
}

//创建Tag
func (d *Dao) CreateTag(c context.Context, tag *model.Tag) (int64, error) {
	if tag == nil || tag.Name == "" {
		return -1, errors.Errorf("d.UpdateTag tag is invalid,tag(%#v)", tag)
	}
	tag.Id = 0

	//看下Tag名是否重复
	e, err := d.CheckExist("mc_tag", "name = ?", tag.Name)
	if err != nil {
		return -1, err
	}
	if e {
		return -1, ecode.UniqueErr
	}

	if err := d.db.Create(tag).Error; err != nil {
		return -1, errors.WithStack(err)
	}
	if err := d.RefreshTagAllCache(c); err != nil {
		return -1, err
	}
	return tag.Id, nil
}

//更新Tag
func (d *Dao) UpdateTag(c context.Context, tag *model.Tag) (int64, error) {
	//先DB再缓存,缓存数据都是不可靠的（可能和数据库的数据不一致），如果需要较为严格的一致性，只以DB的数据为准
	if tag == nil || tag.Id <= 0 || tag.Name == "" {
		return -1, errors.Errorf("d.UpdateTag tag is invalid,tag(%#v)", tag)
	}
	if _, err := d.GetFirstTagFromDB(c, map[string]interface{}{"id": tag.Id}); err != nil {
		if gorm.IsRecordNotFoundError(errors.Cause(err)) {
			return -1, ecode.NothingFound
		}
		return -1, err
	}

	var (
		existTag bool
		err      error
	)
	//看下Tag名是否重复
	existTag, err = d.CheckExist("mc_tag", "name = ? AND id != ?", tag.Name, tag.Id)
	if err != nil {
		return -1, err
	}
	if existTag {
		return -1, ecode.UniqueErr
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
			tx.Commit()
		}
	}()

	//更新Tag表

	tagMaps := map[string]interface{}{
		"name": tag.Name,
	}
	if tag.CreatedAt.Second() > 0 {
		tagMaps["created_at"] = tag.CreatedAt
	}
	if tag.UpdatedAt.Second() > 0 {
		tagMaps["updated_at"] = tag.UpdatedAt
	}

	if err = tx.Table("mc_tag").Where("id=?", tag.Id).Update(tagMaps).Error; err != nil {
		return -1, errors.WithStack(err)
	}
	//更新中间表Name字段
	if err = tx.Table("mc_article_tag").Where("tag_id = ?", tag.Id).
		Update("tag_name", tag.Name).Error; err != nil {
		return -1, errors.WithStack(err)
	}
	//刷新Tag列表缓存
	if err = d.RefreshTagAllCache(c); err != nil {
		return -1, err
	}

	//让关联到的文章刷新tag
	err = d.cacheQueue.Do(c, func(c context.Context) {
		if err := d.RefreshRelateArt(c, tag.Id); err != nil {
			log.SugarWithContext(c).Errorf("d.RefreshRelateArt Err(%#+v)", err)
		}
	})

	return tag.Id, errors.WithStack(err)
}

//删除Tag
func (d *Dao) DeleteTag(c context.Context, tagId int64, physical bool) error {
	if tagId <= 0 {
		return nil
	}
	db := d.db
	if physical {
		db = d.db.Unscoped()
	}
	var err error
	tx := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			if perr := recover(); perr != nil {
				tx.Rollback()
				panic(perr)
			}
			tx.Commit()
		}
	}()

	if err = tx.Where("id=?", tagId).Delete(model.Tag{}).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return errors.WithStack(err)
		}
	}
	if err = tx.Table("mc_article_tag").Where("tag_id = ?", tagId).Delete(model.ArticleTag{}).
		Error; err != nil {
		return errors.WithStack(err)
	}

	if err = d.RefreshTagAllCache(c); err != nil {
		return err
	}
	//让关联到的文章刷新tag
	err = d.cacheQueue.Do(c, func(c context.Context) {
		if err := d.RefreshRelateArt(c, tagId); err != nil {
			log.SugarWithContext(c).Errorf("d.RefreshRelateArt Err(%#+v)", err)
		}
	})

	return errors.WithStack(err)
}

//刷新文章表对应的tag字段
func (d *Dao) RefreshRelateArt(c context.Context, tagId int64) error {
	artTag := make([]*model.ArticleTag, 0)
	if err := d.db.Table("mc_article_tag").Select("article_id").
		Where("tag_id = ?", tagId).Find(&artTag).Error; err != nil {
		return err
	}
	for _, v := range artTag {
		if err := d.RefreshArt(c, v.ArticleId); err != nil {
			return err
		}
	}

	return nil
}

//刷新tag列表缓存
func (d *Dao) RefreshTagAllCache(c context.Context) error {
	res, err := d.GetAllTagFromDB(c)
	if err != nil {
		return err
	}
	if err := d.SetAllTagCache(c, res); err != nil {
		return errors.Wrap(ecode.UpdateCacheErr, err.Error())
	}
	return nil
}

//从DB中直接获取Tag列表
func (d *Dao) GetAllTagFromDB(c context.Context, where ...interface{}) ([]*model.Tag, error) {
	res := make([]*model.Tag, 0)
	err := d.db.Find(&res, where...).Error
	return res, errors.WithStack(err)
}

//从DB中获取单个Tag
func (d *Dao) GetFirstTagFromDB(c context.Context, where ...interface{}) (*model.Tag, error) {
	res := new(model.Tag)
	err := d.db.First(&res, where...).Error
	return res, errors.WithStack(err)
}
