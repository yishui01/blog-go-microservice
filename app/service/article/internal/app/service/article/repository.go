package article

import (
	"blog-go-microservice/app/service/article/internal/app/model/db"
)

type Repository interface {
	Get(id int) (user db.Article, err error)
}
