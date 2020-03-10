package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/utils"

	"github.com/uniplaces/carbon"
	"time"
)

type DBConf struct {
	Driver       string
	Dsn          string
	Active       int
	Idle         int
	IdleTimeout  int
	LogMode      bool
	STable       bool
	TablePrefix  string
	TableHandler func(db *gorm.DB, defaultTableName string) string

	Args []interface{}
}

func NewDB(c *DBConf) *gorm.DB {
	if c == nil {
		c = setDefaultConf()
	}
	db, err := gorm.Open(c.Driver, c.Args...)
	utils.PanicErr(err)

	gorm.DefaultTableNameHandler = c.TableHandler
	db.DB().SetMaxOpenConns(c.Active)
	db.DB().SetMaxIdleConns(c.Idle)
	db.DB().SetConnMaxLifetime(time.Duration(c.IdleTimeout) * time.Hour)
	db.LogMode(c.LogMode)
	db.SingularTable(c.STable)

	//替换掉gorm原来的三个回调方法
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)

	return db
}

func setDefaultConf() *DBConf {
	var conf = new(DBConf)
	if err := viper.Sub("db").Unmarshal(&conf); err != nil {
		panic("unable to decode DBConfig struct, %v" + err.Error())
	}

	conf.TableHandler = func(db *gorm.DB, defaultTableName string) string {
		return conf.TablePrefix + defaultTableName
	}
	conf.Args = []interface{}{conf.Dsn}
	return conf
}

func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := carbon.Now().DateTimeString()

		if createTimeField, ok := scope.FieldByName("CreatedAt"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}
		if modifyTimeField, ok := scope.FieldByName("UpdatedAt"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:updated_at"); !ok {
		scope.SetColumn("updated_at", carbon.Now().DateTimeString())
	}
}

func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedOnField, hasDeletedOnField := scope.FieldByName("DeletedAt")

		if !scope.Search.Unscoped && hasDeletedOnField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deletedOnField.DBName),
				scope.AddToVars(time.Now().Unix()),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}
