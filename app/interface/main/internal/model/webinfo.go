package model

import (
	"net/url"
	"strings"
)

const MUSIC_WEBINFO_KEY = "music"
const FRIEND_WEBINFO_KEY = "friend"
const IMG_WEBINFO_KEY = "backImg"
const PUBLIC_WEBINFO_KEY = "public_conf"
const PRIVATE_WEBINFO_KEY = "private_conf"

type FrontListWebInfo struct {
	PageComm
	Lists []*FrontWebInfoDetail
}

type FrontWebInfoDetail struct {
	Sn     string
	WebKey string
	WebVal string
}

type WebInfo struct {
	Id        int64  `form:"id" json:"id"`
	Sn        string `form:"sn" json:"sn"`
	WebKey    string `form:"web_key" json:"web_key"`
	UniqueVal string `form:"unique_val" json:"unique_val"`
	WebVal    string `form:"web_val" json:"web_val"`
	Status    int32  `form:"status" json:"status"`
	CreatedAt int64  `form:"created_at" json:"created_at"`
	UpdatedAt int64  `form:"updated_at" json:"updated_at"`
	DeletedAt int64  `form:"deleted_at" json:"deleted_at"`
	Ord       string `form:"ord" json:"ord"`
}

func BuildFilter(w url.Values) string {
	maps := map[string]bool{
		"id":         true,
		"sn":         true,
		"web_key":    true,
		"unique_val": true,
		"web_val":    true,
		"status":     true,
		"ord":        true,
	}
	filter := ""
	split := ","
	for k, v := range w {
		if maps[k] {
			filter += split + k + split + v[0]
		}
	}

	return strings.Trim(filter, split)
}
