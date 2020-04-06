package service

import (
	"blog-go-microservice/app/interface/main/internal/model"
	poemPb "blog-go-microservice/app/service/poems/api"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
	"strings"
)

//前台获取诗词
func (s *Service) HomePoemList(c *khttp.Context) {
	req := new(poemPb.PoReq)
	req.PageNum = 1
	req.PageSize = 1
	req.Filter = "cate,huajianji" //前台固定获取花间集的词就行了，因为花间集的词感觉逼格高 ๑乛◡乛๑
	resp, err := s.poemsRPC.Search(c, req)
	if err != nil {
		log.SugarWithContext(c).Errorf("boundary Service.HomePoemList Err(%#+v)", err)
		//挂了就固定返回一句话算了
		c.JSON(s.assignFrontPoems(model.BreakPoemsRes[0]), nil)
		c.Abort()
		return
	}

	c.JSON(s.assignFrontPoems(resp.Lists[0]), nil)

}

func (s *Service) assignFrontPoems(resp *poemPb.PoDetail) *model.FrontPoems {
	detail := new(model.FrontPoems)

	detail.Cate = resp.Cate
	detail.Title = resp.Title
	detail.Author = resp.Author
	detail.Rhythmic = resp.Rhythmic
	detail.Section = resp.Section
	detail.Comment = resp.Comment
	detail.Sn = resp.Sn
	detail.Content = strings.Split(resp.Content, "\n")
	detail.Chapter = strings.Split(resp.Chapter, "\n")
	detail.Paragraphs = strings.Split(resp.Paragraphs, "\n")
	detail.Notes = strings.Split(resp.Notes, "\n")
	return detail
}

/********************************** 后台 ********************************/
func (s *Service) BackPoemList(c *khttp.Context) {
	req := new(poemPb.PoReq)
	if err := c.Bind(req); err != nil {
		return
	}
	resp, err := s.poemsRPC.Search(c, req)
	c.JSON(resp, err)
}
