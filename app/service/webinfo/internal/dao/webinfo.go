package dao

import (
	"blog-go-microservice/app/service/webinfo/internal/model"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strings"
)

var (
	_orderKey = map[string]bool{
		"id":         true,
		"sn":         true,
		"web_key":    true,
		"web_val":    true,
		"status":     true,
		"unique_val": true,
		"ord":        true,
		"created_at": true,
		"deleted_at": true,
	}
)

type Query struct {
	utils.PageRequest
	Filter   string //id,5,sn,ae862,...
	Order    string
	Unscoped bool
}

func (d *Dao) GetInfoListDB(c context.Context, query *Query) ([]*model.WebInfo, error) {
	//这些直接从DB查了，因为要分页、搜索过滤、排序、用缓存不太好存，太麻烦
	var (
		db       = d.db
		err      error
		offset   int64 = 0
		pageSize int64 = 20
	)
	db, err = d.BuildFilter(c, query.Filter, db)
	if err != nil {
		return nil, err
	}

	if query.Order != "" {
		db, err = d.BuildOrder(c, query.Order, db)
		if err != nil {
			return nil, err
		}
	} else {
		db = db.Order("ord desc").Order("id desc")
	}

	if query.Unscoped {
		db = db.Unscoped()
	}

	if query.PageRequest.PageSize > 0 {
		pageSize = int64(query.PageRequest.PageSize)
	}
	//从1开始。1为第一页，2为第二页
	if query.PageRequest.PageNum > 1 {
		offset = (query.PageRequest.PageNum - 1) * pageSize
	}
	db = db.Offset(offset).Limit(pageSize)

	res := make([]*model.WebInfo, 0)
	if err = db.Find(&res).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}

func (d *Dao) CreateInfoDB(c context.Context, info *model.WebInfo) (int64, error) {
	info.Id = 0
	b, err := utils.CheckExist(d.db, "mc_web_info", "web_key=? AND unique_val = ?",
		info.WebKey, info.UniqueVal)
	if err != nil {
		return info.Id, err
	}
	if b {
		return info.Id, ecode.Error(ecode.UniqueErr, info.WebKey+"唯一性错误"+info.UniqueVal)
	}
	if err := d.db.Create(info).Error; err != nil {
		return -1, errors.WithStack(err)
	}
	return info.Id, nil
}

func (d *Dao) UpdateInfoDB(c context.Context, info *model.WebInfo) (int64, error) {
	b, err := utils.CheckExist(d.db, "mc_web_info", "id = ?",
		info.Id)
	if err != nil {
		return info.Id, err
	}
	if !b {
		return info.Id, ecode.NothingFound
	}
	b, err = utils.CheckExist(d.db, "mc_web_info", "id != ? AND web_key=? AND unique_val = ?",
		info.Id, info.WebKey, info.UniqueVal)
	if err != nil {
		return info.Id, err
	}
	if b {
		return info.Id, ecode.Error(ecode.UniqueErr, info.WebKey+"唯一性错误"+info.UniqueVal)
	}

	UpdateMaps := map[string]interface{}{
		"web_key":    info.WebKey,
		"web_val":    info.WebVal,
		"unique_val": info.UniqueVal,
		"ord":        info.Ord,
		"status":     info.Status,
	}
	if info.CreatedAt.Second() > 0 {
		UpdateMaps["created_at"] = info.CreatedAt
	}
	if info.UpdatedAt.Second() > 0 {
		UpdateMaps["updated_at"] = info.UpdatedAt
	}
	if err := d.db.Table("mc_web_info").Where("id=?", info.Id).Update(UpdateMaps).Error; err != nil {
		return info.Id, errors.WithStack(err)
	}
	return info.Id, nil

}

func (d *Dao) DeleteInfoDB(c context.Context, id int64, physical bool) error {
	db := d.db
	if physical {
		db = db.Unscoped()
	}

	if err := db.Where("id=?", id).Delete(&model.WebInfo{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *Dao) BuildFilter(c context.Context, filterStr string, db *gorm.DB) (*gorm.DB, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	if filterStr == "" {
		return db, nil
	}
	filters := strings.Split(filterStr, ",")
	if len(filters)%2 != 0 {
		return nil, ecode.Error(ecode.RequestErr, "filter's params count must be even:"+filterStr)
	}
	columnMaps := map[string][]string{
		"id":         {"id", "="},
		"sn":         {"sn", "="},
		"web_key":    {"web_key", "="},
		"unique_val": {"unique_val", "="},
		"web_val":    {"web_val", "like"},
		"status":     {"status", "="},
		"c_start":    {"created_at", ">="},
		"u_start":    {"updated_at", ">="},
		"d_start":    {"deleted_at", ">="},
		"c_end":      {"created_at", "<="},
		"u_end":      {"updated_at", "<="},
		"d_end":      {"deleted_at", "<="},
	}
	for i := 0; i < len(filters); i += 2 {
		if val, ok := columnMaps[filters[i]]; ok {
			db = db.Where(strings.Join(val, " ")+" ?", filters[i+1])
		}
	}
	return db, nil
}

func (d *Dao) BuildOrder(c context.Context, orderStr string, db *gorm.DB) (*gorm.DB, error) {
	ss := strings.Split(orderStr, ",")
	ordFlag := false
	idFlag := false
	for _, v := range ss {
		if v != "" {
			o := strings.Split(v, "|")
			if len(o) == 2 && _orderKey[o[0]] && (o[1] == "asc" || o[1] == "desc") {
				if o[0] == "ord" {
					ordFlag = true
				}
				if o[0] == "id" {
					idFlag = true
				}
				db = db.Order(o[0] + " " + o[1])
			}
		}
	}
	if !ordFlag {
		db = db.Order("ord desc")
	}
	if !idFlag {
		db = db.Order("id desc")
	}
	return db, nil
}
