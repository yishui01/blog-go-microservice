package dao

import (
	"blog-go-microservice/app/service/article/internal/model"
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

func (d *Dao) EsSearchArt(ctx context.Context, req *model.ArtQueryReq) (*elastic.SearchResult, error) {
	query := elastic.NewBoolQuery()

	if req.KeyWords != "" {
		keyWordQuery := elastic.NewMultiMatchQuery(req.KeyWords).
			FieldWithBoost("title", 3).
			FieldWithBoost("content", 2)
		query = query.Must(keyWordQuery)
	}
	if req.Tags != "" {
		tags := make([]interface{}, len(req.Tags))
		for k, v := range strings.Split(req.Tags, ",") {
			tags[k] = v
		}
		q := elastic.NewTermsQuery("tags", tags...)
		query = query.Filter(q)
	}

	if req.Status != -1 {
		q := elastic.NewTermsQuery("status", req.Status)
		query = query.Filter(q)
	}

	if req.CreatedAt > 0 {
		q := elastic.NewRangeQuery("created_at").Gte(req.CreatedAt)
		query = query.Filter(q)
	}

	if req.UpdatedAt > 0 {
		q := elastic.NewRangeQuery("updated_at").Lte(req.UpdatedAt)
		query = query.Filter(q)
	}

	if !req.Unscoped {
		q := elastic.NewTermQuery("deleted_at", time.Time{})
		query = query.Filter(q)
	}
	if log.LogConf.Level == zap.DebugLevel.String() {
		source, err := query.Source()
		log.SugarLogger.Debugf("EsSearchArt Source:%#v\n Err:%#v\n", source, err)
	}
	search := d.es.Search(model.ART_ES_INDEX).Query(query).
		From(int((req.PageNum - 1) * req.PageSize)).
		Size(int(req.PageSize))

	if req.Order == "" {
		search.Sort("id", false)
	} else {
	}

	res, err := search.Do(ctx)
	if err != nil {
		source, err := query.Source()
		log.SugarLogger.Errorf("EsSearchArt Search Err: Source:%#v\n Err:%#v\n", source, err)
	}
	log.SugarLogger.Debugf("EsSearchArt Res:%#v\n", res)
	if res != nil {
		log.SugarLogger.Debugf("EsSearchArt Res.TotalHits:%#v\n", res.TotalHits())
		log.SugarLogger.Debugf("EsSearchArt Res.Hits.Hits:%#v\n", res.Hits.Hits)
	}

	return res, errors.Wrap(err, "Elastic search art Err")
}

func (d *Dao) EsPutArt(ctx context.Context, art *model.Article) (*elastic.IndexResponse, error) {
	exists, err := d.es.IndexExists(model.ART_ES_INDEX).Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Exists Article Index Err")
	}
	if !exists {
		// Create a new index.
		createIndex, err := d.es.CreateIndex(model.ART_ES_INDEX).BodyString(model.Mapping).Do(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "create "+model.ART_ES_INDEX+" Index Err")
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
			return nil, errors.New("create " + model.ART_ES_INDEX + " Index Acknowledged is false")
		}
	}

	resp, err := d.es.Index().Index(model.ART_ES_INDEX).
		Id(strconv.Itoa(int(art.Id))).BodyJson(art.ToEsMap()).Do(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "insert into es err:")
	}
	return resp, nil
}

func (d *Dao) EsDeleteArt(ctx context.Context, id int64) (*elastic.DeleteResponse, error) {
	exists, err := d.es.IndexExists(model.ART_ES_INDEX).Do(ctx)
	if err != nil || !exists {
		return nil, errors.Wrap(err, "Exists "+model.ART_ES_INDEX+" Index Err On EsDeleteArt")
	}
	resp, err := d.es.Delete().Index(model.ART_ES_INDEX).
		Id(strconv.Itoa(int(id))).Do(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "EsDeleteArt es err:")
	}
	return resp, nil
}
