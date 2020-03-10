package model

import (
	"github.com/jinzhu/gorm"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strings"
)

type Article struct {
	Id        int64  `gorm:"primary_key";json:"_"`
	Sn        string `json:"sn"`
	Title     string `json:"title"`
	Tags      string `json:"tags"`
	Img       string `json:"img"`
	Content   string `json:"content"`
	Status    int64  `json:"status"`
	CreatedAt string `json:"created_at"` //time.Time
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at" gorm:"default:''"`
}

type ArtQueryReq struct {
	utils.PageRequest
	KeyWords string
	Tags     []string
	Status   int64
	Order    []string
	Debug    bool
	Unscoped bool //true时会查找已经软删除的记录
}

func (user *Article) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("Sn", utils.Md5ByTime("key"))

	return nil
}

func (self *Article) ToEsMap() map[string]interface{} {
	maps := make(map[string]interface{})
	maps["Id"] = self.Id
	maps["Sn"] = self.Sn
	maps["Title"] = self.Title
	maps["Tags"] = strings.Split(self.Tags, ",")
	maps["Img"] = self.Img
	maps["Content"] = self.Content
	maps["Status"] = self.Status
	maps["CreatedAt"] = self.CreatedAt
	maps["UpdatedAt"] = self.UpdatedAt
	maps["DeletedAt"] = self.DeletedAt
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
					"type":"nested"
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
