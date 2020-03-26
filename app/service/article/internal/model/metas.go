package model

type Metas struct {
	ArticleId int64  `gorm:"primary_key";json:"article_id"`
	Sn        string `json:"sn"`
	ViewCount int64  `json:"view_count"`
	CmCount   int64  `json:"cm_count"`
	LaudCount int64  `json:"laud_count"`
}

const ArtIdRedisKey = "art_id"
const ViewRedisKey = "view_count"
const CmRedisKey = "cm_count"
const LaudRedisKey = "laud_count"
