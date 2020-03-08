package service

import (
	pb "blog-go-microservice/app/service/article/api"
	"blog-go-microservice/app/service/article/internal/dao"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
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
	//art, err := s.dao.GetArtBySn(artReq.Sn)
	//if err != nil {
	//	return nil, err
	//}
	//reply = new(pb.ArtDetailResp)
	//reply.Sn = art.Sn
	//reply.Title = art.Title
	//reply.Img = art.Img
	//reply.Content = art.Content
	//reply.CreatedAt = art.CreatedAt.Format("2006-01-02 15:04:05")
	//reply.UpdatedAt = art.UpdatedAt.Format("2006-01-02 15:04:05")
	//reply.DeletedAt = art.DeletedAt.Format("2006-01-02 15:04:05")
	reply = new(pb.ArtDetailResp)
	reply.Sn = "777"
	reply.Title = "999"
	return reply, nil
}

func (s *Service) GetArtStream(artReq pb.Article_GetArtStreamServer) (err error) {
	return nil
}

func (s *Service) Close() {
}

func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
