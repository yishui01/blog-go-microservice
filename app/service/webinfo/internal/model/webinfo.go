package model

import (
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"time"
)

const WEB_CACHE_KEY = "web_cache_key_"

type WebInfo struct {
	Id        int64
	Sn        string
	WebKey    string
	UniqueVal string
	WebVal    string
	Status    int32
	Ord       string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"-"` //gorm ignore in create and update method
}

func (user *WebInfo) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("Sn", utils.SubMd5ByTime("key3f61t651cxv 6.135~@!#%^&$^f13w32`**/3568yr8b532.=/?3"))
	return nil
}

func WebCacheKey(key string) string {
	return WEB_CACHE_KEY + key
}
