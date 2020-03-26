package model

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strings"
	"time"
)

type Article struct {
	Id        int64     `gorm:"primary_key";json:"_"`
	Sn        string    `json:"sn"`
	Title     string    `json:"title"`
	Tags      string    `json:"tags"`
	Img       string    `json:"img"`
	Content   string    `json:"content"`
	Status    int32     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"-"` //gorm ignore
}

//从ES查出来后解析json到结构体用的
type EsArticle struct {
	Id        int64     `gorm:"primary_key";json:"_"`
	Sn        string    `json:"sn"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags"`
	Img       string    `json:"img"`
	Content   string    `json:"content"`
	Status    int32     `json:"status"`
	ViewCount int64     `json:"view_count"`
	CmCount   int64     `json:"cm_count"`
	LaudCount int64     `json:"laud_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"-"` //gorm ignore
}

type ArtQueryReq struct {
	utils.PageRequest
	KeyWords  string
	Tags      string
	Status    int32
	Order     string
	CreatedAt int64
	UpdatedAt int64
	Unscoped  bool //true时会查找已经软删除的记录
}

func (user *Article) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("Sn", utils.SubMd5ByTime("key34648-+-/`124bes***R*3568yr8b532.=/?3"))
	return nil
}

func ArtToEsMap(ctx context.Context, art *Article, metas *Metas) map[string]interface{} {
	maps := make(map[string]interface{})
	if art != nil {
		maps["id"] = art.Id
		maps["sn"] = art.Sn
		maps["title"] = art.Title
		maps["tags"] = strings.Split(art.Tags, ",")
		maps["img"] = art.Img
		maps["content"] = art.Content
		maps["status"] = art.Status
		maps["created_at"] = art.CreatedAt
		maps["updated_at"] = art.UpdatedAt
		maps["deleted_at"] = art.DeletedAt
	}
	if metas != nil {
		maps["view_count"] = metas.ViewCount
		maps["laud_count"] = metas.LaudCount
		maps["cm_count"] = metas.CmCount
	}
	log.SugarWithContext(ctx).Debugf("Article ToEsMap:%#v\n", maps)
	return maps
}

func ArtLockKey(sn string) string {
	return "art_lock_" + sn
}

const ART_ES_INDEX = "article"

//number_of_shards  是数据分片数，默认为5，有时候设置为3
//number_of_replicas 是数据备份数，如果只有一台机器，设置为0
const Mapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
			"properties":{
				"id":{
					"type":"integer"
				},
				"sn":{
					"type":"keyword"
				},
				"title":{
					"type":"text"
				},
				"tags":{
					"type":"keyword"
				},
				"img":{
					"type":"keyword"
				},
				"content":{
					"type":"text"
				},
				"status":{
					"type":"integer"
				},
				"image":{
					"type":"keyword"
				},
				"view_count":{
					"type":"integer"
				},
				"laud_count":{
					"type":"integer"
				},
				"cm_count":{
					"type":"integer"
				},
				"created_at":{
					"type":"date"
				},
				"updated_at":{
					"type":"date"
				},
				"deleted_at":{
					"type":"date"
				}
			}
		}
}`
