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
	CreatedAt time.Time `json:"created_at" gorm:"-"`
	UpdatedAt time.Time `json:"updated_at" gorm:"-"`
	DeletedAt time.Time `json:"deleted_at" gorm:"-"`
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
	CreatedAt time.Time `json:"created_at" gorm:"-"`
	UpdatedAt time.Time `json:"updated_at" gorm:"-"`
	DeletedAt time.Time `json:"deleted_at" gorm:"-"`
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

func (self *Article) ToEsMap(ctx context.Context) map[string]interface{} {
	maps := make(map[string]interface{})
	maps["id"] = self.Id
	maps["sn"] = self.Sn
	maps["title"] = self.Title
	maps["tags"] = strings.Split(self.Tags, ",")
	maps["img"] = self.Img
	maps["content"] = self.Content
	maps["status"] = self.Status
	maps["created_at"] = self.CreatedAt
	maps["updated_at"] = self.UpdatedAt
	maps["deleted_at"] = self.DeletedAt
	log.SugarWithContext(ctx).Debugf("ToEsMap:%#v\n", maps)
	return maps
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
