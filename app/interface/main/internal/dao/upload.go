package dao

import (
	"blog-go-microservice/app/interface/main/internal/model"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"path/filepath"
)

/**********************************  公共  ************************************************************/
//创建文件记录
func (d *Dao) UploadCreate(c context.Context, fileName, key, url string) error {
	upload := model.Upload{
		Name:      fileName,
		Url:       url,
		Uploadkey: key,
		Tag:       filepath.Ext(key),
	}
	return errors.WithStack(d.db.Create(&upload).Error)
}

/***********************************  后台  ************************************************************/
//文件列表
func (d *Dao) BackUploadList(c context.Context, params *model.BackUploadQuery) ([]*model.Upload, int, error) {
	query := d.db
	if params.Name != "" {
		query = query.Where("name like ?", "%"+params.Name+"%")
	}
	if params.Tag != "" {
		query = query.Where("tag = ?", params.Tag)
	}
	uploads := make([]*model.Upload, 0)
	total := 0
	if err := query.Model(&model.Upload{}).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "get total err")
	}
	err := query.Order("id desc").Offset((params.PageNum - 1) * params.PageSize).Limit(params.PageSize).Find(&uploads).Error
	return uploads, total, errors.WithStack(err)
}

//查找单个文件
func (d *Dao) UploadFind(c context.Context, id int) (*model.Upload, error) {
	var (
		err error
	)
	f := &model.Upload{}
	if err = d.db.Where("id=?", id).First(&f).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ecode.NothingFound
		}
		return nil, errors.WithStack(err)
	}
	return f, nil
}

//删除单个文件
func (d *Dao) UploadDelete(c context.Context, id int) error {
	if err := d.db.Where("id=?", id).Delete(model.Upload{}).Error; err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithStack(err)
	}
	return nil
}
