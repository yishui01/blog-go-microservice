package model

import (
	"github.com/zuiqiangqishao/framework/pkg/utils"
)

const POEMS_ES_INDEX  = "poems"

type Poem struct {
	Id         int64 `json:"id,string"`
	Cate       string
	Title      string
	Author     string
	Content    string
	Chapter    string
	Paragraphs string
	Notes      string
	Rhythmic   string
	Section    string
	Comment    string
	Sn         string
	CreateTime string
}

type Query struct {
	utils.PageRequest
	Filter string
}
