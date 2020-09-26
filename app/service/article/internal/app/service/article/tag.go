package article

import (
	pb "blog-go-microservice/app/service/article/pb"
	"blog-go-microservice/app/service/article/internal/app/model/db"
	"context"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strconv"
)

func (s *Service) TagList(c context.Context, req *pb.TagListReq) (*pb.TagListResp, error) {
	tags, err := s.dao.GetTagAll(c)
	if err != nil {
		log.SugarWithContext(c).Errorf("s.dao.GetTagAll Err:(%#+v)", err)
		return nil, err
	}
	reply := new(pb.TagListResp)
	for _, v := range tags {
		if req.Keyword != "" && req.Keyword != v.Name {
			continue
		}
		reply.Lists = append(reply.Lists, &pb.TagDetailResp{Id: v.Id, Name: v.Name, CreatedAt: v.CreatedAt.Unix(), UpdatedAt: v.UpdatedAt.Unix(), DeletedAt: v.DeletedAt.Unix()})
	}
	return reply, nil
}

func (s *Service) CreateTag(c context.Context, req *pb.SaveTagReq) (*pb.SaveResp, error) {
	return s.saveTag(c, req, false)
}

func (s *Service) UpdateTag(c context.Context, req *pb.SaveTagReq) (*pb.SaveResp, error) {
	return s.saveTag(c, req, true)
}

func (s *Service) saveTag(c context.Context, req *pb.SaveTagReq, isUpdate bool) (*pb.SaveResp, error) {
	var (
		err   error
		tagId int64
	)
	reply := new(pb.SaveResp)
	tag := AssignTag(req)
	if isUpdate {
		tagId, err = s.dao.UpdateTag(c, tag)
	} else {
		tagId, err = s.dao.CreateTag(c, tag)
	}
	if err != nil {
		if _, ok := errors.Cause(err).(ecode.Codes); ok {
			return nil, err
		}
		log.SugarWithContext(c).Errorf("s.dao.saveTag tag(%#+v), req(%#+v), Err:(%#+v)", tag, req, err)
		return nil, ecode.Error(ecode.ServerErr, err.Error())
	}
	reply.Data = strconv.FormatInt(tagId, 10)
	err = s.dao.RefreshTagAllCache(c)
	if err != nil {
		log.SugarWithContext(c).Errorf("s.dao.RefreshTagAllCache tag(%#+v), req(%#+v), Err:(%#+v)", tag, req, err)
		reply.Status = 1 //flag cache err
	}
	return reply, nil
}

func (s *Service) DeleteTag(c context.Context, req *pb.DelTagReq) (*pb.SaveResp, error) {
	res := new(pb.SaveResp)
	var err error
	if err = s.dao.DeleteTag(c, req.Id, req.Physical); err != nil {
		log.SugarWithContext(c).Errorf("Service s.dao.DeleteTag req:(%#+v), Err:(%+v)", req, err)
	}
	return res, err
}

func AssignTag(req *pb.SaveTagReq) *db.Tag {
	tag := new(db.Tag)
	tag.Id = req.Id
	tag.Name = req.Name
	tag.CreatedAt = utils.TimeUnixToTime(req.CreatedAt)
	tag.UpdatedAt = utils.TimeUnixToTime(req.UpdatedAt)
	return tag
}
