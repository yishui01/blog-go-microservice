package service

import (
	pb "blog-go-microservice/app/service/article/api"
	"blog-go-microservice/app/service/article/internal/dao"
	"context"
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

func (s *Service) GetArtBySn(ctx context.Context, artReq *pb.ArtDetailRequest) (reply *pb.ArtDetailResp, err error) {
	art, err := s.dao.GetArtBySn(artReq.Sn)
	if err != nil {
		return nil, err
	}
	reply = new(pb.ArtDetailResp)
	reply.Sn = art.Sn
	reply.Title = art.Title
	reply.Img = art.Img
	reply.Content = art.Content
	reply.CreatedAt = art.CreatedAt.Format("2006-01-02 15:04:05")
	reply.UpdatedAt = art.UpdatedAt.Format("2006-01-02 15:04:05")
	reply.DeletedAt = art.DeletedAt.Format("2006-01-02 15:04:05")
	return reply, nil
}

func (s *Service) Close() {
	s.dao.Close()
}

func (s *Service) Ping(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	//todo
	return nil, nil
}