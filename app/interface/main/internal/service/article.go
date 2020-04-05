package service

import (
	pbArt "blog-go-microservice/app/service/article/api"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
)

func (s *Service) HomeArtList(c *khttp.Context) {
	//文章列表
	req := new(pbArt.ArtListRequest)
	resp, err := s.artRPC.ArtList(c, req)
	c.JSON(resp, err)
}

func (s *Service) HomeArtDetail(c *khttp.Context) {

}

/********************************** 后台 ********************************/
func (s *Service) BackArtList(c *khttp.Context) {

}

func (s *Service) BackArtDetail(c *khttp.Context) {

}

func (s *Service) BackArtCreate(c *khttp.Context) {

}

func (s *Service) BackArtUpdate(c *khttp.Context) {

}

func (s *Service) BackArtDelete(c *khttp.Context) {

}
