package model

type Metas struct {
	ArticleId int64
	Sn        string
	ViewCount int64
	CmCount   int64
	LaudCount int64
}

const ArtIdRedisKey = "art_id"
const ViewRedisKey = "view_count"
const CmRedisKey = "cm_count"
const LaudRedisKey = "laud_count"
