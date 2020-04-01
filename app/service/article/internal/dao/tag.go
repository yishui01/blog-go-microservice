package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strconv"
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
	err = d.db.Table("mc_tag").Find(&res).Error
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
	e, err := utils.CheckExist(d.db, "mc_tag", "name = ?", tag.Name)
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
func (d *Dao) UpdateTag(c context.Context, tag *model.Tag) (tagId int64, err error) {
	//先DB再缓存,缓存数据都是不可靠的（可能和数据库的数据不一致），如果需要较为严格的一致性，只以DB的数据为准
	if tag == nil || tag.Id <= 0 || tag.Name == "" {
		return -1, errors.Errorf("d.UpdateTag tag is invalid,tag(%#v)", tag)
	}
	if _, err := d.GetFirstTagFromDB(c, map[string]interface{}{"id": tag.Id}); err != nil {
		if gorm.IsRecordNotFoundError(errors.Cause(err)) {
			return tag.Id, ecode.NothingFound
		}
		return tag.Id, err
	}

	var (
		existTag bool
	)
	//看下Tag名是否重复
	existTag, err = utils.CheckExist(d.db, "mc_tag", "name = ? AND id != ?", tag.Name, tag.Id)
	if err != nil {
		return tag.Id, err
	}
	if existTag {
		return tag.Id, ecode.UniqueErr
	}

	tx := d.db.Begin()

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
		tx.Rollback()
		return tag.Id, errors.WithStack(err)
	}
	//更新中间表Name字段
	if err = tx.Table("mc_article_tag").Where("tag_id = ?", tag.Id).
		Update("tag_name", tag.Name).Error; err != nil {
		tx.Rollback()
		return tag.Id, errors.WithStack(err)
	}

	//这里要先commit，成功后再刷新缓存，不然刷新缓存的时候，从db里面不能读到未提交的事务
	if err = tx.Commit().Error; err != nil {
		return tag.Id, errors.Wrap(err, "d.UpdateTag transaction commit error")
	}

	/***************************************非事务部分*************************************/
	//获取关联到的文章ID
	relateArtIds, err := d.GetRelateArtIds(c, tag.Id)
	if err != nil {
		return tag.Id, err
	}

	//让关联到的文章刷新tag
	if err := d.RefreshRelateArt(c, relateArtIds); err != nil {
		return tag.Id, errors.WithStack(err)
	}

	//刷新Tag列表缓存
	if err = d.RefreshTagAllCache(c); err != nil {
		return tag.Id, err
	}

	return tag.Id, errors.WithStack(err)
}

//先找出tag所有的关联的文章id
func (d *Dao) GetRelateArtIds(c context.Context, tagId int64) (artIds []int64, err error) {
	rows, err := d.db.Raw("SELECT article_id FROM mc_article_tag WHERE tag_id = ?", tagId).Rows()
	if err != nil {
		fmt.Println("8888888888888", err)
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	artId := 0
	relateArtIds := []int64{}
	for rows.Next() {
		if err := rows.Scan(&artId); err != nil {
			return nil, err
		}
		if artId > 0 {
			relateArtIds = append(relateArtIds, int64(artId))
		}
	}
	return relateArtIds, nil
}

//删除Tag
func (d *Dao) DeleteTag(c context.Context, tagId int64, physical bool) (err error) {
	if tagId <= 0 {
		return errors.Wrap(ecode.RequestErr, "tagId must grater than 0,now is"+strconv.FormatInt(tagId, 10))
	}
	db := d.db
	if physical {
		db = d.db.Unscoped()
	}
	tx := db.Begin()

	//先删除Tag表
	if err = tx.Where("id=?", tagId).Delete(model.Tag{}).Error; err != nil {
		tx.Rollback()
		if err != gorm.ErrRecordNotFound {
			return errors.WithStack(err)
		}
	}
	//获取关联到的文章ID
	relateArtIds, err := d.GetRelateArtIds(c, tagId)
	if err != nil {
		tx.Rollback()
		return err
	}

	//删除中间关联表
	if err = tx.Table("mc_article_tag").Where("tag_id = ?", tagId).Delete(model.ArticleTag{}).
		Error; err != nil {
		tx.Rollback()
		return errors.WithStack(err)
	}

	//这里要先commit，成功后再刷新缓存，不然刷新缓存的时候，从db里面不能读到未提交的事务
	if err = tx.Commit().Error; err != nil {
		return errors.Wrap(err, "d.DeleteTag transaction commit error")
	}

	/********************************非事务部分*****************************************/
	//刷新文章冗余tag字段就不走事务了，走事务调用链太长，报错了直接看log手动调整
	if err := d.RefreshRelateArt(c, relateArtIds); err != nil {
		return err
	}
	if err = d.RefreshTagAllCache(c); err != nil {
		return err
	}

	return errors.WithStack(err)
}

//刷新文章表对应的tag字段
func (d *Dao) RefreshRelateArt(c context.Context, ArtIds []int64) error {
	for _, v := range ArtIds {
		if err := d.RefreshArt(c, v); err != nil {
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
