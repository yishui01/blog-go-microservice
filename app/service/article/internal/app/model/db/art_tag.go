package db

import (
	"strings"
	"time"
)

type ArticleTag struct {
	ArticleId int64
	TagId     int64
	TagName   string
	DeletedAt time.Time `json:"deleted_at" gorm:"-"` //gorm ignore
}

func BuildArtTagStr(s []*ArticleTag) string {
	t := []string{}
	for _, v := range s {
		t = append(t, v.TagName)
	}
	return strings.Join(t, ",")
}
