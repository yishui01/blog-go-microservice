package service

import (
	"blog-go-microservice/app/interface/main/internal/dao"
	"blog-go-microservice/app/interface/main/internal/model"
	"blog-go-microservice/app/interface/main/middleware/auth"
	artCli "blog-go-microservice/app/service/article/api"
	poemsCli "blog-go-microservice/app/service/poems/api"
	webInfoCli "blog-go-microservice/app/service/webinfo/api"
	"github.com/spf13/viper"
)

type Service struct {
	d             *dao.Dao
	artRPC        artCli.ArticleClient
	poemsRPC      poemsCli.PoemsClient
	webInfoRPC    webInfoCli.WebInfoClient
	jwt           *auth.JWTCfg
	Permis        *auth.Auth
	UploadManager *model.UploadManager
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
	permis := auth.New(nil, d)
	jwt := auth.LoadJWTConfInFile()
	qiniu := &model.Qiniu{}
	if err := viper.Sub("qiniu").Unmarshal(&qiniu); err != nil {
		panic("qiniu config parse err:" + err.Error())
	}
	return &Service{
		d:             d,
		artRPC:        artRPC,
		poemsRPC:      poemsRPC,
		webInfoRPC:    webInfoRPC,
		jwt:           jwt,
		Permis:        permis,
		UploadManager: model.NewUploadManager(qiniu),
	}
}

func (s *Service) Close() {
	s.d.Close()
}
