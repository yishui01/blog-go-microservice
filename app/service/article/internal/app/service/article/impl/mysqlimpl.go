package impl

import (
	"blog-go-microservice/app/service/article/internal/app/model/db"
	"blog-go-microservice/app/service/article/internal/app/service/article"
	"github.com/douyu/jupiter/pkg/store/gorm"
)

type mysqlImpl struct {
	gh *gorm.DB
}

// NewMysqlImpl construct an instance of mysqlImpl
func NewMysqlImpl(gh *gorm.DB) article.Repository {
	return &mysqlImpl{
		gh: gh,
	}
}
func (m *mysqlImpl) Get(id int) (user db.Article, err error) {
	user = db.Article{}
	err = m.gh.Where("id = ?", id).Find(&user).Error
	return
}
