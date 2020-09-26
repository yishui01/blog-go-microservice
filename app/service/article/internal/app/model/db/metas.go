package db

import (
	"context"
	"github.com/zuiqiangqishao/framework/pkg/log"
)

type Metas struct {
	ArticleId int64  `gorm:"primary_key";json:"article_id"`
	Sn        string `json:"sn"`
	ViewCount int64  `json:"view_count"`
	CmCount   int64  `json:"cm_count"`
	LaudCount int64  `json:"laud_count"`
}

//metas hash field
const ArtIdRedisKey = "art_id"
const ViewRedisKey = "view_count"
const CmRedisKey = "cm_count"
const LaudRedisKey = "laud_count"

func (self *Metas) ToEsMap(ctx context.Context) map[string]interface{} {
	maps := make(map[string]interface{})
	maps["article_id"] = self.ArticleId
	maps["sn"] = self.Sn
	maps["view_count"] = self.ViewCount
	maps["cm_count"] = self.CmCount
	maps["laud_count"] = self.LaudCount

	log.SugarWithContext(ctx).Debugf("Metas ToEsMap:%#v\n", maps)
	return maps
}
