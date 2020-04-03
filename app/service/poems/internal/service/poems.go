package service

import (
	pb "blog-go-microservice/app/service/poems/api"
	"blog-go-microservice/app/service/poems/internal/dao"
	"blog-go-microservice/app/service/poems/internal/model"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
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

func (s *Service) Search(c context.Context, req *pb.PoReq) (*pb.PoResp, error) {
	var (
		res = new(pb.PoResp)
		err error
	)
	lists, total, err := s.dao.EsSearch(c, AssignQuery(req))
	if err != nil {
		if _, ok := errors.Cause(err).(ecode.Codes); ok {
			return nil, err
		}
		log.SugarWithContext(c).Errorf("s.dao.EsSearch req(%#v),ERR(%#+v)", req, err)
		return res, ecode.ServerErr
	}
	details := make([]*pb.PoDetail, 0)
	for _, info := range lists {
		if info == nil || info.Id <= 0 {
			continue
		}
		details = append(details, AssignDetail(info))
	}
	res.Lists = details
	res.Page = req.PageNum
	res.Total = total

	return res, nil
}

func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
func (s *Service) Close() {

}

func AssignQuery(req *pb.PoReq) *model.Query {
	q := new(model.Query)
	q.PageSize = req.PageSize
	q.PageNum = req.PageNum
	q.Filter = req.Filter
	return q
}

func AssignDetail(info *model.Poem) *pb.PoDetail {
	if info == nil {
		return nil
	}

	de := new(pb.PoDetail)
	de.Id = info.Id
	de.Cate = info.Cate
	de.Title = info.Title
	de.Author = info.Author
	de.Content = info.Content
	de.Chapter = info.Chapter
	de.Paragraphs = info.Paragraphs
	de.Notes = info.Notes
	de.Rhythmic = info.Rhythmic
	de.Section = info.Section
	de.Comment = info.Comment
	de.Sn = info.Sn
	de.CreateTime = info.CreateTime
	return de
}
