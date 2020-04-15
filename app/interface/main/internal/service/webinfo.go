package service

import (
	"blog-go-microservice/app/interface/main/internal/model"
	webPb "blog-go-microservice/app/service/webinfo/api"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
	"strconv"
)

func (s *Service) HomeWebInfoList(c *khttp.Context) {
	req := s.assignWebListReq(c)
	rpcRes, err := s.webInfoRPC.InfoList(c, req)
	res := new(model.FrontListWebInfo)
	if err != nil {
		log.SugarWithContext(c).Errorf("boundary service.HomeWebInfoList Err:(%#+v)", err)
		c.JSON(res, ecode.OK)
		c.Abort()
		return
	}

	lists := make([]*model.FrontWebInfoDetail, len(rpcRes.Lists))
	for k, v := range rpcRes.Lists {
		lists[k] = s.assignWebDetail(v)
	}

	res.Lists = lists
	res.Total = rpcRes.Total
	res.PageSize = rpcRes.Size_
	res.Page = rpcRes.Page
	c.JSON(res, nil)
}

/********************************** 后台 ********************************/
func (s *Service) BackWebInfoList(c *khttp.Context) {
	req := new(webPb.InfoReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	req.Filter = model.BuildFilter(c.Request.Form)
	resp, err := s.webInfoRPC.InfoList(c, req)
	c.JSON(resp, err)
}

func (s *Service) BackWebInfoCreate(c *khttp.Context) {
	req := new(webPb.SaveInfoReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	resp, err := s.webInfoRPC.CreateInfo(c, req)
	c.JSON(resp, err)
}

func (s *Service) BackWebInfoUpdate(c *khttp.Context) {
	req := new(webPb.SaveInfoReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	resp, err := s.webInfoRPC.UpdateInfo(c, req)
	c.JSON(resp, err)
}

func (s *Service) BackWebInfoDelete(c *khttp.Context) {
	req := new(webPb.DelInfoReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	resp, err := s.webInfoRPC.DeleteInfo(c, req)
	c.JSON(resp, err)
}

/**********************************参数赋值*****************************/
func (s *Service) assignWebListReq(c *khttp.Context) *webPb.InfoReq {
	var (
		page    int64 = 1
		size    int32 = 20
		keyMaps       = map[string]bool{
			model.MUSIC_WEBINFO_KEY:  true,
			model.FRIEND_WEBINFO_KEY: true,
			model.IMG_WEBINFO_KEY:    true,
			model.PUBLIC_WEBINFO_KEY: true,
		}
	)

	req := new(webPb.InfoReq)

	params := c.Request.Form
	filterStr := "web_key,"
	web_key := params.Get("web_key")
	if web_key == "" || !keyMaps[web_key] {
		web_key = "music"
	}
	filterStr += web_key
	formPage, err := strconv.Atoi(params.Get("page_num"))
	if err == nil && formPage >= 1 && formPage < 100000 {
		page = int64(formPage)
	}
	formSize, err := strconv.Atoi(params.Get("page_size"))
	if err == nil && formSize >= 1 && formSize < 100000 {
		size = int32(formSize)
	}
	params.Get("")
	req.Filter = filterStr
	req.PageNum = page
	req.PageSize = size
	return req
}

func (s *Service) assignWebDetail(detail *webPb.InfoDetail) *model.FrontWebInfoDetail {
	res := new(model.FrontWebInfoDetail)
	res.WebKey = detail.WebKey
	res.WebVal = detail.WebVal
	res.Sn = detail.Sn
	return res
}
