package service

import (
	pb "blog-go-microservice/app/service/article/api"
	"blog-go-microservice/app/service/article/internal/dao"
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strconv"
	"strings"
)

type Service struct {
	dao *dao.Dao
}

func New(d *dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		dao: d,
	}
	cf = s.Close
	return
}

func (s *Service) ArtList(ctx context.Context, listReq *pb.ArtListRequest) (*pb.ArtListResp, error) {
	query := new(model.ArtQueryReq)
	query.PageNum = listReq.PageNum
	query.PageSize = listReq.PageSize
	query.KeyWords = listReq.Keyword
	query.Tags = listReq.Tags
	query.Status = listReq.Status
	query.Order = listReq.Order
	query.CreatedAt = listReq.CreatedAt
	query.UpdatedAt = listReq.UpdatedAt
	query.Unscoped = listReq.Unscoped
	esArtList, err := s.dao.ArtList(ctx, query)
	if err != nil {
		log.SugarLogger.Errorf("Service ArtList Err:(%+v)", err)
	}
	reply := new(pb.ArtListResp)
	reply.Total = int64(len(esArtList))
	t := make([]*pb.ArtDetailResp, len(esArtList))
	for k, v := range esArtList {
		t[k] = &pb.ArtDetailResp{
			Id:        v.Id,
			Sn:        v.Sn,
			Title:     v.Title,
			Tags:      strings.Join(v.Tags, ","),
			Status:    v.Status,
			Img:       v.Img,
			Content:   v.Content,
			CreatedAt: v.CreatedAt.Unix(),
			UpdatedAt: v.UpdatedAt.Unix(),
			DeletedAt: v.DeletedAt.Unix(),
		}
	}
	reply.Lists = t
	reply.Page = listReq.PageNum
	reply.Size_ = listReq.PageSize

	return reply, err
}

func (s *Service) GetArtBySn(ctx context.Context, artReq *pb.ArtDetailRequest) (*pb.ArtDetailResp, error) {
	art, err := s.dao.GetArtBySn(ctx, artReq.Sn)
	if err != nil {
		log.SugarLogger.Errorf("Service GetArtBySn Err:(%+v)", err)
		return nil, err
	}
	reply := new(pb.ArtDetailResp)
	reply.Id = art.Id
	reply.Sn = art.Sn
	reply.Tags = art.Tags
	reply.Title = art.Title
	reply.Img = art.Img
	reply.Content = art.Content
	reply.CreatedAt = art.CreatedAt.Unix()
	reply.UpdatedAt = art.UpdatedAt.Unix()
	return reply, nil
}

func (s *Service) SaveArt(ctx context.Context, req *pb.SaveArtReq) (*pb.SaveArtResp, error) {
	art := new(model.Article)
	art.Id = req.Id
	art.Tags = req.Tags
	art.Title = req.Title
	art.Img = req.Img
	art.Content = req.Content
	art.Status = req.Status
	art.CreatedAt = utils.TimeUnixToTime(req.CreatedAt)
	art.CreatedAt = utils.TimeUnixToTime(req.CreatedAt)
	art, err := s.dao.SaveArt(ctx, art)
	if err != nil {
		log.SugarLogger.Errorf("Service CreateArt  Err:(%+v)", err)
		return nil, err
	}
	reply := new(pb.SaveArtResp)
	var bs []byte
	if bs, err = utils.JsonMarshal(map[string]string{
		"id": strconv.Itoa(int(art.Id)),
		"sn": art.Sn,
	}); err != nil {
		log.SugarLogger.Errorf("Service JsonEncode id and sn  Err:(%+v)", err)
		return nil, err
	}

	reply.Data = string(bs)
	return reply, err
}

func (s *Service) DeleteArt(ctx context.Context, req *pb.DelArtRequest) (*pb.SaveArtResp, error) {
	res := new(pb.SaveArtResp)
	if err := s.dao.DelArt(ctx, req.Id, req.Physical); err != nil {
		log.SugarLogger.Errorf("Service DeleteArt  Err:(%+v)", err)
		res.Status = 1
		res.Msg = err.Error()
	}
	return res, nil
}

func (s *Service) Close() {
}

func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
