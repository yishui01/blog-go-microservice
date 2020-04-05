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

func (d *Dao) UserLogin(c context.Context, username, passwd string) (*model.User, error) {
	users := new(model.User)
	if err := d.db.Raw(_findUserName_sql, username).Scan(&users).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ecode.Unauthorized
		}
		return nil, errors.WithStack(err)
	}
	if utils.ValidatePassword(passwd, users.PassWord) != nil {
		//success
		return users, nil
	}
	return nil, ecode.Unauthorized
}

func (d *Dao) UserCreate(c context.Context, user *model.User, isAdmin bool) (userID int64, err error) {
	user.ID = 0
	if isAdmin {
		user.ISSuper = 1
	} else {
		user.ISSuper = 0
	}
	if err := d.db.Create(user).Error; err != nil {
		return 0, errors.WithStack(err)
	}
	return user.ID, nil
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

func (d *Dao) BackUserUpdate(c context.Context, user *model.User) (userId int64, err error) {
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
		"username": user.UserName,
		"nickname": user.NickName,
		"avatar":   user.Avatar,
		"desc":     user.Desc,
		"email":    user.Email,
		"phone":    user.Phone,
		"is_super": user.ISSuper,
		"status":   user.Status,
	}).Error
	return user.ID, errors.WithStack(err)
}

func (d *Dao) UserUpdatePass(c context.Context, userSn string, oldPass, newPass string) error {
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

func (d *Dao) BackUserDelete(c context.Context, sn string) error {
	if err := d.db.Table("mc_user").Where("sn=?", sn).Delete(model.User{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ecode.RequestErr
		}
		return errors.WithStack(err)
	}
	return nil
}
