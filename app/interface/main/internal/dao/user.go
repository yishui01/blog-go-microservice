package dao

import (
	"blog-go-microservice/app/interface/main/internal/model"
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/utils"
)

var _findUserName_sql = "SELECT * FROM `mc_user` WHERE `username` = ?"

/************************************公共*************************************/
func (d *Dao) UserLogin(c context.Context, username, passwd string) (*model.User, error) {
	users := new(model.User)
	if err := d.db.Raw(_findUserName_sql, username).Scan(&users).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ecode.RequestErr
		}
		return nil, errors.WithStack(err)
	}
	if utils.ValidatePassword(passwd, users.PassWord) == nil {
		if users.Status != 0 {
			return nil, ecode.AccessDenied
		}
		//success
		return users, nil
	}
	return nil, ecode.RequestErr
}

func (d *Dao) FindFirstUser(c context.Context, query string, args ...interface{}) (*model.User, error) {
	user := new(model.User)
	if err := d.db.Table("mc_user").Where(query, args...).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ecode.NothingFound
		}
		return nil, errors.WithStack(err)
	}
	return user, nil
}

/**********************************   前台   ************************************************/
func (d *Dao) FrontUserUpdate(c context.Context, user *model.User) (userId int64, err error) {
	var (
		exists bool
	)
	if exists, err = utils.CheckExist(d.db, "mc_user", "sn=?", user.Sn); err != nil {
		return 0, err
	}
	if !exists {
		return 0, ecode.NothingFound
	}
	err = d.db.Table("mc_user").Where("sn", user.Sn).Update(map[string]interface{}{
		"nickname": user.NickName,
		"avatar":   user.Avatar,
		"desc":     user.Desc,
	}).Error
	return user.ID, errors.WithStack(err)
}

/*************************************   后台   **********************************************/
func (d *Dao) BackUserList(c context.Context, params *model.BackUserQuery) ([]*model.User, int, error) {
	query := d.db
	if params.UserName != "" {
		query = query.Where("username like ?", "%"+params.UserName+"%")
	}
	if params.NickName != "" {
		query = query.Where("nickname like ?", "%"+params.NickName+"%")
	}
	if params.Cate != "" {
		query = query.Where("cate = ? ", params.Cate)
	}
	if params.Status != "" {
		query = query.Where("status = ? ", params.Status)
	}
	if params.ISSuper != "" {
		query = query.Where("is_super = ?", params.ISSuper)
	}
	if params.OpenCate != "" {
		query = query.Where("open_cate = ?", params.OpenCate)
	}
	if params.ISDelete == "1" {
		query = query.Where("deleted_at IS NOT NULL")
	}
	if params.CreatedAt != "" {
		query = query.Where("created_at >= ? ", params.CreatedAt)
	}
	if params.UpdatedAt != "" {
		query = query.Where("updated_at >= ? ", params.UpdatedAt)
	}
	users := make([]*model.User, 0)
	total := 0
	if err := query.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "get total err")
	}
	err := query.Offset((params.PageNum - 1) * params.PageSize).Limit(params.PageSize).Find(&users).Error
	return users, total, errors.WithStack(err)
}

func (d *Dao) BackUserCreate(c context.Context, user *model.User, isAdmin bool) (userID int64, err error) {
	user.ID = 0
	if isAdmin {
		user.ISSuper = 1
	} else {
		user.ISSuper = 0
	}
	//查看username是否已经存在
	var exists bool
	if exists, err = utils.CheckExist(d.db.Unscoped(), "mc_user", "username=?", user.UserName); err != nil {
		return 0, err
	}
	if exists {
		return 0, ecode.UniqueErr
	}
	encPass, err := utils.GeneratePassword(user.PassWord)
	if err != nil {
		return 0, err
	}
	user.PassWord = string(encPass)
	if err := d.db.Create(user).Error; err != nil {
		return 0, errors.WithStack(err)
	}
	return user.ID, nil
}

func (d *Dao) BackUserUpdate(c context.Context, user *model.User) (userId int64, err error) {
	var (
		exists bool
	)
	//检查用户是否存在
	if exists, err = utils.CheckExist(d.db, "mc_user", "id=?", user.ID); err != nil {
		return 0, err
	}
	if !exists {
		return 0, ecode.NothingFound
	}

	//检查修改后的名称是否重复
	if exists, err = utils.CheckExist(d.db, "mc_user", "username=? AND id != ?", user.UserName, user.ID); err != nil {
		return 0, err
	}
	if exists {
		return 0, ecode.UniqueErr
	}

	updateData := map[string]interface{}{
		"nickname": user.NickName,
		"avatar":   user.Avatar,
		"desc":     user.Desc,
		"email":    user.Email,
		"phone":    user.Phone,
		"is_super": user.ISSuper,
		"status":   user.Status,
	}
	if user.PassWord != "" {
		updateData["password"], err = utils.GeneratePassword(user.PassWord)
		if err != nil {
			return 0, err
		}
	}

	err = d.db.Table("mc_user").Where("id = ?", user.ID).Update(updateData).Error
	return user.ID, errors.WithStack(err)
}

func (d *Dao) CheckOldPass(c context.Context, userSn string, oldPass string) error {
	var (
		err  error
		user model.User
	)

	if err = d.db.Table("mc_user").Where("sn=?", userSn).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ecode.NothingFound
		}
		return errors.WithStack(err)
	}
	if utils.ValidatePassword(oldPass, user.PassWord) != nil {
		return ecode.RequestErr
	}
	return nil
}

func (d *Dao) UserUpdatePass(c context.Context, userSn string, newPass string) error {
	newDBPassWord, err := utils.GeneratePassword(newPass)
	newPassToken := utils.GetUUID()
	if err != nil {
		return err
	}
	if err := d.db.Exec("UPDATE mc_user SET password= ?,password_token=? WHERE sn = ?",
		newDBPassWord, newPassToken, userSn).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *Dao) BackUserDelete(c context.Context, id uint) error {
	if err := d.db.Table("mc_user").Where("id=?", id).Delete(model.User{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ecode.RequestErr
		}
		return errors.WithStack(err)
	}
	return nil
}
