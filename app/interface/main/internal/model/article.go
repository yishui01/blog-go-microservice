package model

import (
	pb "blog-go-microservice/app/service/article/api"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strings"
)

type FontListArticle struct {
	PageComm
	Lists []*FrontArtDetail
}

type FrontArtDetail struct {
	Sn        string   `json:"sn"`
	Title     string   `json:"title"`
	Img       string   `json:"img"`
	Tags      []string `json:"tags"`
	Content   string   `json:"content"`
	ViewCount int64    `json:"view_count"`
	CmCount   int64    `json:"cm_count"`
	LaudCount int64    `json:"laud_count"`
	CreatedAt string   `json:"created_at"`
}

func AssignFrontArticle(resp *pb.ArtDetailResp) *FrontArtDetail {
	font := new(FrontArtDetail)
	font.Sn = resp.Sn
	font.Title = resp.Title

	font.Tags = strings.Split(resp.Tags, ",")
	font.Content = resp.Content
	font.CreatedAt = utils.TimeUnixToTime(resp.CreatedAt).Format(TIME_LAYOUT)
	font.ViewCount = resp.ViewCount
	font.CmCount = resp.CmCount
	font.LaudCount = resp.LaudCount
	return font
}
