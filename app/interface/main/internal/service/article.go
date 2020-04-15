package service

import (
	"blog-go-microservice/app/interface/main/internal/model"
	pbArt "blog-go-microservice/app/service/article/api"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	khttp "github.com/zuiqiangqishao/framework/pkg/net/http"
	"github.com/zuiqiangqishao/framework/pkg/utils/business"
	"strconv"
	"strings"
	"time"
)

/********************************************** 前台 ***********************************************************/
func (s *Service) HomeArtList(c *khttp.Context) {
	//文章列表
	req := s.assignListReq(c)
	rpcRes, err := s.artRPC.ArtList(c, req)
	res := new(model.FontListArticle)
	if rpcRes != nil {
		lists := make([]*model.FrontArtDetail, len(rpcRes.Lists))
		for k, v := range rpcRes.Lists {
			lists[k] = model.AssignFrontArticle(v)
		}
		res.Lists = lists
		res.Total = rpcRes.Total
		res.Page = rpcRes.Page
		res.PageSize = rpcRes.Size_
	}
	c.JSON(res, err)
}

//文章详情
func (s *Service) HomeArtDetail(c *khttp.Context) {
	sn := c.Request.Form.Get("sn")
	if sn == "" {
		c.JSON("请传入文章标识", ecode.RequestErr)
		c.Abort()
		return
	}
	req := new(pbArt.ArtDetailRequest)
	req.Sn = sn
	resp, err := s.artRPC.GetArtBySn(c, req)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	c.JSON(model.AssignFrontArticle(resp), nil)
}

/****************************************** 后台管理系统api **************************************************************/
//后台请求直接透传给grpc调用，根据pb文件的struct标签进行表单验证
//开始想着将后台http请求转发到grpc-gateway的http端口，然而没有grpc-gateway服务地址的获取机制，实际上grpc-gateway的http地址是已经
//注册到etcd中的，只是服务发现的逻辑被内置在grpc的resolve中了，没有暴露出来，如果这里再调用Build去Resolver手动获取地址，再过滤出http地址
//也是可以完成的，只是很麻烦，不如直接用rpc client一样，还是长连接
func (s *Service) BackArtList(c *khttp.Context) {
	req := new(pbArt.ArtListRequest)
	//这里status默认是0，这样会有问题，不知道这个0是查的值=0还是没有传status参数,因此和rpc server端约定好 -1 代表不过滤status
	req.Status = -1
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.ArtList(c, req)
	c.JSON(rpcRes, err)
}

func (s *Service) BackArtDetail(c *khttp.Context) {
	req := new(pbArt.ArtDetailRequest)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.GetArtBySn(c, req)
	c.JSON(rpcRes, err)
}

func (s *Service) BackArtCreate(c *khttp.Context) {
	req := new(pbArt.SaveArtReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.CreateArt(c, req)
	c.JSON(rpcRes, err)
}

func (s *Service) BackArtUpdate(c *khttp.Context) {
	req := new(pbArt.SaveArtReq)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.UpdateArt(c, req)
	c.JSON(rpcRes, err)
}

func (s *Service) BackArtDelete(c *khttp.Context) {
	req := new(pbArt.DelArtRequest)
	if err := c.MustBind(req); err != nil {
		return
	}
	rpcRes, err := s.artRPC.DeleteArt(c, req)
	c.JSON(rpcRes, err)
}

/******************************************  参数赋值  ********************************************************/
func (s *Service) assignListReq(c *khttp.Context) *pbArt.ArtListRequest {
	var (
		page int64 = 1
		size int32 = 10
	)

	req := new(pbArt.ArtListRequest)
	params := c.Request.Form

	formPage, err := strconv.Atoi(params.Get("page_num"))
	if err == nil && formPage >= 1 && formPage < 100000 {
		page = int64(formPage)
	}
	formSize, err := strconv.Atoi(params.Get("page_size"))
	if err == nil && formSize >= 1 && formSize < 100000 {
		size = int32(formSize)
	}

	formOrder := strings.Trim(params.Get("order"), " ")
	matchSlice := business.ArtOrderReg().FindStringSubmatch(formOrder)
	if len(matchSlice) >= 3 && business.ArtOrderKey()[matchSlice[1]] && (matchSlice[2] == "asc" || matchSlice[2] == "desc") {
		req.Order = formOrder
	}
	createdAt, err := time.Parse(model.TIME_LAYOUT, params.Get("created_at"))
	if err == nil {
		req.CreatedAt = createdAt.Unix()
	}
	req.Tags = strings.Trim(params.Get("tags"), " ")
	req.Keyword = strings.Trim(params.Get("Keyword"), " ")
	req.PageNum = page
	req.PageSize = size
	req.Status = 1 //只看已上线的

	return req
}
