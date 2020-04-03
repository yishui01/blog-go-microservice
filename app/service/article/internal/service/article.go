package service

import (
	pb "blog-go-microservice/app/service/article/api"
	"blog-go-microservice/app/service/article/internal/dao"
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/sync/errgroup"
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

//文章列表
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
	query.Terms = listReq.Terms
	esArtList, total, err := s.dao.ArtMetasList(ctx, query)
	reply := new(pb.ArtListResp)
	if err != nil {
		if _, ok := errors.Cause(err).(ecode.Codes); ok {
			return nil, err
		}
		log.SugarWithContext(ctx).Errorf("s.dao.ArtList Query(%#+v),  Err:(%#+v)", query, err)
		return reply, ecode.ServerErr
	}

	reply.Total = total
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
			CmCount:   v.CmCount,
			LaudCount: v.LaudCount,
			ViewCount: v.ViewCount,
			CreatedAt: v.CreatedAt.Unix(),
			UpdatedAt: v.UpdatedAt.Unix(),
			DeletedAt: v.DeletedAt.Unix(),
		}

	}
	reply.Lists = t
	reply.Page = listReq.PageNum
	reply.Size_ = listReq.PageSize

	return reply, nil
}

//文章详情（包含metas)
func (s *Service) GetArtBySn(ctx context.Context, artReq *pb.ArtDetailRequest) (*pb.ArtDetailResp, error) {
	var (
		art      *model.Article
		metas    *model.Metas
		artErr   error
		metasErr error
	)
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) error {
		art, artErr = s.dao.GetArtBySn(ctx, artReq.Sn)
		if artErr != nil {
			if ecode.EqualError(ecode.NothingFound, artErr) {
				artErr = ecode.NothingFound
			}
			log.SugarWithContext(ctx).Errorf("s.dao.GetArtBySn  artReq：(%#+v),Err:(%#+v)", artReq, artErr)
			artErr = ecode.ServerErr
		}
		return nil
	})

	g.Go(func(ctx context.Context) error {
		metas, metasErr = s.dao.GetMetasBySn(ctx, artReq.Sn)
		if metasErr != nil {
			log.SugarWithContext(ctx).Errorf("s.dao.GetMetasBySn  artReq：(%#+v),Err:(%#+v)", artReq, metasErr)
		}
		if metas == nil {
			metas = new(model.Metas)
		}
		return nil
	})
	if se := s.dao.AddMetasCount(ctx, artReq.Sn, model.ViewRedisKey); se != nil {
		log.SugarWithContext(ctx).Warnf("d.AddMetasCount Err(%#v)", se)
	}
	if err := g.Wait(); err != nil {
		log.SugarWithContext(ctx).Errorf("s.dao.GetMetasBySn wait Err artReq：(%#+v),Err:(%s)", artReq, err.Error())
		return nil, ecode.ServerErr
	}

	if artErr != nil {
		return nil, artErr
	}
	if art == nil {
		return nil, ecode.NothingFound
	}

	reply := AssignDetail(art, metas)

	return reply, nil
}

func AssignDetail(art *model.Article, metas *model.Metas) *pb.ArtDetailResp {
	reply := new(pb.ArtDetailResp)
	reply.Id = art.Id
	reply.Sn = art.Sn
	reply.Tags = art.Tags
	reply.Title = art.Title
	reply.Img = art.Img
	reply.Content = art.Content
	reply.CreatedAt = art.CreatedAt.Unix()
	reply.UpdatedAt = art.UpdatedAt.Unix()
	reply.LaudCount = metas.LaudCount
	reply.CmCount = metas.CmCount
	reply.ViewCount = metas.ViewCount
	return reply
}

func AssignSave(req *pb.SaveArtReq) (*model.Article, *model.Metas) {
	art := new(model.Article)
	art.Id = req.Id
	art.Tags = req.Tags
	art.Title = req.Title
	art.Img = req.Img
	art.Content = req.Content
	art.Status = req.Status
	art.CreatedAt = utils.TimeUnixToTime(req.CreatedAt)
	art.UpdatedAt = utils.TimeUnixToTime(req.UpdatedAt)
	metas := new(model.Metas)
	metas.Sn = req.Sn
	metas.ArticleId = req.Id
	metas.CmCount = req.CmCount
	metas.ViewCount = req.ViewCount
	metas.LaudCount = req.LaudCount
	return art, metas
}

//创建文章
func (s *Service) CreateArt(ctx context.Context, req *pb.SaveArtReq) (*pb.SaveResp, error) {
	return s.SaveArt(ctx, req, false)
}

//修改文章
func (s *Service) UpdateArt(ctx context.Context, req *pb.SaveArtReq) (*pb.SaveResp, error) {
	return s.SaveArt(ctx, req, true)
}

func (s *Service) SaveArt(ctx context.Context, req *pb.SaveArtReq, isUpdate bool) (*pb.SaveResp, error) {
	var (
		err   error
		artId int64
	)
	var reply = new(pb.SaveResp)

	art, metas := AssignSave(req)
	if isUpdate {
		artId, err = s.dao.UpdateArtMetas(ctx, art, metas)
	} else {
		artId, err = s.dao.CreateArtMetas(ctx, art, metas)
	}
	if err != nil {
		if _, ok := errors.Cause(err).(ecode.Codes); ok {
			return nil, err
		}
		log.SugarWithContext(ctx).Errorf("s.SaveArt art(%#+v), metas(%#+v), Err:(%#+v)", art, metas, err)
		return nil, ecode.Error(ecode.ServerErr, err.Error())
	}
	reply.Data = strconv.FormatInt(artId, 10)

	err = s.dao.RefreshArt(ctx, art.Id)
	reply.Data = strconv.FormatInt(artId, 10)
	if err != nil {
		log.SugarWithContext(ctx).Errorf("s.dao.SetArtCache art(%#+v), metas(%#+v), Err:(%#+v)", art, metas, err)
		reply.Status = 1 //flag cache err
	}

	return reply, nil
}

//删除文章
func (s *Service) DeleteArt(ctx context.Context, req *pb.DelArtRequest) (*pb.SaveResp, error) {
	res := new(pb.SaveResp)
	var err error
	if err := s.dao.DelArtMetas(ctx, req.Id, req.Physical); err != nil {
		log.SugarWithContext(ctx).Errorf("Service s.dao.DelArt req:(%#+v), Err:(%+v)", req, err)
	}
	return res, err
}

func (s *Service) Close() {
}

func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
