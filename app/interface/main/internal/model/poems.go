package model

import pb "blog-go-microservice/app/service/poems/api"

type FrontPoems struct {
	Cate       string
	Title      string
	Author     string
	Content    []string
	Chapter    []string
	Paragraphs []string
	Notes      []string
	Rhythmic   string
	Section    string
	Comment    string
	Sn         string
}

var (
	BreakPoemsRes = []*pb.PoDetail{
		{
			Id:         20441,
			Cate:       "huajianji",
			Title:      "浣溪沙·钿匣菱花锦带垂",
			Author:     "薛昭蕴",
			Content:    "",
			Chapter:    "",
			Paragraphs: "钿匣菱花锦带垂，静临栏槛卸头时，约鬟低珥算归期。\n茂苑草青湘渚阔，梦余空有漏依依，二年终日损芳菲。",
			Notes:      "1.菱花--菱花镜。\n2.卸头时--卸妆时。\n3.珥--冠上的垂珠。\n4.茂苑--茂苑：在今江 苏吴县太湖北。湘渚：湘水中的小洲。",
			Rhythmic:   "浣溪沙",
			Section:    "",
			Comment:    "",
			Sn:         "babe7afbfd8c418a1353d4c3",
		},
	}
)
