package dao

import (
	"blog-go-microservice/app/service/article/internal/app/model/db"
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils/business"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

//查询ES中的文章数据
func (d *Dao) EsSearchArtMetas(ctx context.Context, req *db.ArtQueryReq) (*elastic.SearchResult, error) {
	query := elastic.NewBoolQuery()
	if req.KeyWords != "" {
		keyWordQuery := elastic.NewMultiMatchQuery(req.KeyWords).
			FieldWithBoost("title", 3).
			FieldWithBoost("content", 2)
		query = query.Must(keyWordQuery)
	}
	if req.Tags != "" {
		tags := make([]interface{}, 0)
		for _, v := range strings.Split(req.Tags, ",") {
			if strings.Trim(v, " ") != "" {
				tags = append(tags, v)
			}
		}
		q := elastic.NewTermsQuery("tags", tags...)
		query = query.Filter(q)
	}

	if req.Status != -1 {
		q := elastic.NewTermsQuery("status", req.Status)
		query = query.Filter(q)
	}

	if req.Terms != "" {
		ts := strings.Split(req.Terms, ",")
		if len(ts)%2 != 0 {
			return nil, ecode.Error(ecode.RequestErr, "terms参数个数必须为双数")
		}
		for i := 0; i < len(ts); i += 2 {
			q := elastic.NewTermQuery(ts[i], ts[i+1])
			query = query.Filter(q)
		}

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
	if log.GetLogConf().Level == zap.DebugLevel.String() {
		source, err := query.Source()
		log.SugarWithContext(ctx).Debugf("EsSearchArt Source:%#v\n Err:%#v\n", source, err)
	}
	search := d.es.Search(db.ART_ES_INDEX).Query(query).
		From(int((req.PageNum - 1) * int64(req.PageSize))).
		Size(int(req.PageSize))

	//created_at|desc
	if req.Order == "" {
		search.Sort("id", false)
	} else {
		matchSlice := business.ArtOrderReg().FindStringSubmatch(req.Order)
		if len(matchSlice) >= 3 && business.ArtOrderKey()[matchSlice[1]] && (matchSlice[2] == "asc" || matchSlice[2] == "desc") {
			search.Sort(matchSlice[1], matchSlice[2] == "asc")
		} else {
			search.Sort("id", false)
		}
	}

	res, err := search.Do(ctx)
	if err != nil {
		source, _ := query.Source()
		log.SugarWithContext(ctx).Errorf("EsSearchArt Search Err: Source:%#v\n Err:%#v\n", source, err)
	}
	log.SugarWithContext(ctx).Debugf("EsSearchArt Res:%#v\n", res)
	if res != nil {
		log.SugarWithContext(ctx).Debugf("EsSearchArt Res.TotalHits:%#v\n", res.TotalHits())
		log.SugarWithContext(ctx).Debugf("EsSearchArt Res.Hits.Hits:%#v\n", res.Hits.Hits)
	}

	return res, errors.Wrap(err, "Elastic search art Err")
}

//更新整个文档
func (d *Dao) EsPutArtMetas(ctx context.Context, art *db.Article, metas *db.Metas) (*elastic.IndexResponse, error) {
	exists, err := d.es.IndexExists(db.ART_ES_INDEX).Do(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !exists {
		// Create a new index.
		createIndex, err := d.es.CreateIndex(db.ART_ES_INDEX).BodyString(db.Mapping).Do(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
			return nil, errors.New("create " + db.ART_ES_INDEX + " Index Acknowledged is false")
		}
	}

	resp, err := d.es.Index().Index(db.ART_ES_INDEX).
		Id(strconv.Itoa(int(art.Id))).BodyJson(db.ArtToEsMap(ctx, art, metas)).Do(context.TODO())

	if err != nil {
		return resp, errors.WithStack(err)
	}
	return resp, nil
}

//更新文档部分字段
func (d *Dao) EsUpdateArtMetas(ctx context.Context, art *db.Article, metas *db.Metas) (*elastic.UpdateResponse, error) {
	var id int64
	if art != nil {
		id = art.Id
	}
	if id == 0 && metas != nil {
		id = metas.ArticleId
	}
	if id == 0 {
		return nil, errors.New("not set invalid ID")
	}
	resp, err := d.es.Update().Index(db.ART_ES_INDEX).Id(strconv.FormatInt(id, 10)).
		Doc(db.ArtToEsMap(ctx, art, metas)).Do(ctx)
	if err != nil {
		return resp, errors.WithStack(err)
	}
	return resp, nil
}

func (d *Dao) EsDeleteArtMetas(ctx context.Context, id int64) (*elastic.DeleteResponse, error) {
	exists, err := d.es.IndexExists(db.ART_ES_INDEX).Do(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !exists {
		return nil, nil
	}

	resp, err := d.es.Delete().Index(db.ART_ES_INDEX).Id(strconv.Itoa(int(id))).Do(ctx)
	if err != nil {
		if elastic.IsNotFound(err) {
			return resp, nil
		}
		return nil, errors.WithStack(err)
	}
	return resp, nil
}
