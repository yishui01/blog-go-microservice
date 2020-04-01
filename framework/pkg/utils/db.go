package utils

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

func CheckExist(db *gorm.DB, tableName string, query interface{}, args ...interface{}) (bool, error) {
	if db == nil {
		return false, errors.New("db is nil")
	}
	count := 0
	if err := db.Table(tableName).Where(query, args...).Count(&count).Error; err != nil {
		return false, errors.WithStack(err)
	}
	return count > 0, nil
}
