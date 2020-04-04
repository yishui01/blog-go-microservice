package service

import (
	"blog-go-microservice/app/interface/main/internal/dao"
	artCli "blog-go-microservice/app/service/article/api"
	poemsCli "blog-go-microservice/app/service/poems/api"
	webInfoCli "blog-go-microservice/app/service/webinfo/api"
)

type Service struct {
	d          *dao.Dao
	artRPC     artCli.ArticleClient
	poemsRPC   poemsCli.PoemsClient
	webInfoRPC webInfoCli.WebInfoClient
}

func New() *Service {
	artRPC, err := artCli.NewClient(nil)
	if err != nil {
		panic(err)
	}

	poemsRPC, err := poemsCli.NewClient(nil)
	if err != nil {
		panic(err)
	}
	webInfoRPC, err := webInfoCli.NewClient(nil)
	if err != nil {
		panic(err)
	}
	d, _ := dao.New()
	return &Service{
		d:          d,
		artRPC:     artRPC,
		poemsRPC:   poemsRPC,
		webInfoRPC: webInfoRPC,
	}
}

func (s *Service) Close() {
	s.d.Close()
}
