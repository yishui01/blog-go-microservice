package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"strconv"
)

func (d *Dao) PutArtToEs(ctx context.Context, art *model.Article) (*elastic.IndexResponse, error) {
	exists, err := d.es.IndexExists(model.ART_ES_INDEX).Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Exists Article Index Err")
	}
	if !exists {
		// Create a new index.
		createIndex, err := d.es.CreateIndex(model.ART_ES_INDEX).BodyString(model.Mapping).Do(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "create Article Index Err")
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
			return nil, errors.New("create Article Index Acknowledged is false")
		}
	}

	resp, err := d.es.Index().Index(model.ART_ES_INDEX).
		Id(strconv.Itoa(int(art.Id))).BodyJson(art.ToEsMap()).Do(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "insert into es err:")
	}
	return resp, nil
}
