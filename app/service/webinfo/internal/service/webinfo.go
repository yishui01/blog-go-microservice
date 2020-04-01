package service

import (
	pb "blog-go-microservice/app/service/webinfo/api"
	"blog-go-microservice/app/service/webinfo/internal/dao"
	"blog-go-microservice/app/service/webinfo/internal/model"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"strconv"
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

func (s *Service) InfoList(c context.Context, req *pb.InfoReq) (*pb.InfoResp, error) {
	res := new(pb.InfoResp)
	lists, err := s.dao.GetInfoListDB(c, AssignQuery(req))
	if err != nil {
		if _, ok := errors.Cause(err).(ecode.Codes); ok {
			return nil, err
		}
		log.SugarWithContext(c).Errorf("s.dao.GetInfoList req(%#v),ERR(%#+v)", req, err)
		return res, ecode.ServerErr
	}
	details := make([]*pb.InfoDetail, 0)
	for _, info := range lists {
		if info == nil || info.Id <= 0 {
			continue
		}
		details = append(details, AssignDetail(info))
	}
	res.Lists = details
	res.Page = req.PageNum
	res.Total = int64(len(lists))

	return res, nil
}
func (s *Service) CreateInfo(c context.Context, req *pb.SaveInfoReq) (*pb.SaveResp, error) {
	return s.saveInfo(c, req, false)
}
func (s *Service) UpdateInfo(c context.Context, req *pb.SaveInfoReq) (*pb.SaveResp, error) {
	return s.saveInfo(c, req, true)
}
func (s *Service) saveInfo(c context.Context, req *pb.SaveInfoReq, isUpdate bool) (*pb.SaveResp, error) {
	var (
		reply  = new(pb.SaveResp)
		infoId int64
		err    error
	)
	info := AssignSaveReq(req)
	if !isUpdate {
		infoId, err = s.dao.CreateInfoDB(c, info)
	} else {
		infoId, err = s.dao.UpdateInfoDB(c, info)
	}
	if err != nil {
		if _, ok := errors.Cause(err).(ecode.Codes); ok {
			return nil, err
		}
		log.SugarWithContext(c).Errorf("s.saveInfo req(%#v),Err(%#+v)", req, err)
		return nil, ecode.Error(ecode.ServerErr, err.Error())
	}
	reply.Data = strconv.FormatInt(infoId, 10)

	return reply, nil

}
func (s *Service) DeleteInfo(c context.Context, req *pb.DelInfoReq) (*pb.SaveResp, error) {
	var (
		reply = new(pb.SaveResp)
		err   error
	)
	err = s.dao.DeleteInfoDB(c, req.Id, req.Physical)
	if err != nil {
		log.SugarWithContext(c).Errorf("Service s.dao.DeleteInfo req:(%#+v), Err:(%+v)", req, err)
	}
	return reply, nil
}

func (s *Service) Close() {
}

func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func AssignQuery(req *pb.InfoReq) *dao.Query {
	q := new(dao.Query)
	q.PageSize = req.PageSize
	q.PageNum = req.PageNum
	q.Order = req.Order
	q.Filter = req.Filter
	q.Unscoped = req.Unscoped
	return q
}

func AssignDetail(info *model.WebInfo) *pb.InfoDetail {
	if info == nil {
		return nil
	}
	de := new(pb.InfoDetail)
	de.Id = info.Id
	de.Sn = info.Sn
	de.WebKey = info.WebKey
	de.UniqueVal = info.UniqueVal
	de.WebVal = info.WebVal
	de.Status = info.Status
	de.Ord = info.Ord
	de.CreatedAt = info.CreatedAt.Unix()
	de.UpdatedAt = info.UpdatedAt.Unix()
	de.DeletedAt = info.DeletedAt.Unix()
	return de
}

func AssignSaveReq(req *pb.SaveInfoReq) *model.WebInfo {
	if req == nil {
		return nil
	}
	info := new(model.WebInfo)
	info.Id = req.Id
	info.WebKey = req.WebKey
	info.WebVal = req.WebVal
	info.UniqueVal = req.UniqueVal
	info.Status = req.Status
	info.Ord = req.Ord
	info.Sn = req.Sn
	info.CreatedAt = utils.TimeUnixToTime(req.CreatedAt)
	info.UpdatedAt = utils.TimeUnixToTime(req.UpdatedAt)
	info.DeletedAt = utils.TimeUnixToTime(req.DeletedAt)
	return info
}
