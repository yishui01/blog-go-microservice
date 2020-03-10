package service

import (
	pb "blog-go-microservice/app/service/article/api"
	"blog-go-microservice/app/service/article/internal/dao"
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/zuiqiangqishao/framework/pkg/log"
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
	query.KeyWords = listReq.Keyword
	query.PageNum = listReq.Page
	query.PageSize = listReq.Size_

	res, err := s.dao.ArtList(ctx, query)
	if err != nil {
		log.SugarLogger.Errorf("Service ArtList Err:(%+v)", err)
	}

	reply := new(pb.ArtListResp)
	reply.Total = res.Total
	t := make([]*pb.ArtDetailResp, len(res.Lists))
	for k, v := range res.Lists {
		t[k] = &pb.ArtDetailResp{
			Id:        v.Id,
			Sn:        v.Sn,
			Title:     v.Title,
			Tags:      v.Tags,
			Status:    v.Status,
			Img:       v.Img,
			Content:   v.Content,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			DeletedAt: v.DeletedAt,
		}
	}
	reply.Lists = t
	reply.Page = res.Page
	reply.Size_ = res.Size

	return reply, err
}

func (s *Service) GetArtBySn(ctx context.Context, artReq *pb.ArtDetailRequest) (*pb.ArtDetailResp, error) {
	art, err := s.dao.GetArtBySn(ctx, artReq.Sn)
	if err != nil {
		log.SugarLogger.Errorf("Service GetArtBySn Err:(%+v)", err)
		return nil, err
	}
	reply := new(pb.ArtDetailResp)
	reply.Sn = art.Sn
	reply.Tags = art.Tags
	reply.Title = art.Title
	reply.Img = art.Img
	reply.Content = art.Content
	reply.CreatedAt = art.CreatedAt
	reply.UpdatedAt = art.UpdatedAt
	return reply, nil
}

func (s *Service) SaveArt(ctx context.Context, req *pb.SaveArtReq) (*pb.SaveArtResp, error) {
	art := new(model.Article)
	art.Id = req.Id
	art.Sn = req.Sn
	art.Tags = req.Tags
	art.Title = req.Title
	art.Img = req.Img
	art.Content = req.Content
	art.Status = req.Status
	art, err := s.dao.SaveArt(ctx, art)
	if err != nil {
		log.SugarLogger.Errorf("Service CreateArt  Err:(%+v)", err)
		return nil, err
	}

	reply := new(pb.SaveArtResp)
	return reply, err

}

func (s *Service) DeleteArt(ctx context.Context, req *pb.DelArtRequest) (*pb.SaveArtResp, error) {
	return nil, nil
}

func (s *Service) Close() {
}

func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
