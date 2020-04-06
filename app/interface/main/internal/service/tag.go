package service

import (
	"blog-go-microservice/app/interface/main/internal/model"
	pb "blog-go-microservice/app/service/article/api"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
)

func (s *Service) HomeTagList(c *khttp.Context) {
	req := new(pb.TagListReq)
	rpcRes, err := s.artRPC.TagList(c, req)
	if err != nil {
		log.SugarWithContext(c).Errorf("boundary.service.HomeTagList err(%#+v)", err)
		c.JSON(nil, ecode.ServerErr)
		c.Abort()
		return
	}
	lists := make([]*model.FrontTagDetail, len(rpcRes.Lists))
	if rpcRes != nil {
		for k, v := range rpcRes.Lists {
			lists[k] = &model.FrontTagDetail{Name: v.Name}
		}
	}
	c.JSON(lists, err)

}

/********************************** 后台 ********************************/
func (s *Service) BackTagList(c *khttp.Context) {
	req := new(pb.TagListReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.TagList(c, req)
	c.JSON(rpcRes, err)
}

func (s *Service) BackTagCreate(c *khttp.Context) {
	req := new(pb.SaveTagReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.CreateTag(c, req)
	c.JSON(rpcRes, err)
}

func (s *Service) BackTagUpdate(c *khttp.Context) {
	req := new(pb.SaveTagReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.UpdateTag(c, req)
	c.JSON(rpcRes, err)
}

func (s *Service) BackTagDelete(c *khttp.Context) {
	req := new(pb.DelTagReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.DeleteTag(c, req)
	c.JSON(rpcRes, err)
}
