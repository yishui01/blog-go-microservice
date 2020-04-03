package dao

import (
	"blog-go-microservice/app/service/poems/internal/model"
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/zuiqiangqishao/framework/pkg/ecode"
	"github.com/zuiqiangqishao/framework/pkg/log"
	"github.com/zuiqiangqishao/framework/pkg/utils"
	"go.uber.org/zap"
	"strings"
)

var (
	_filterMaps = map[string]string{
		"sn":     "term",
		"id":     "term",
		"cate":   "term",
		"author": "term",

		"title":      "match",
		"content":    "match",
		"chapter":    "match",
		"paragraphs": "match",
		"notes":      "match",
		"rhythmic":   "match",
		"section":    "match",
		"comment":    "match",
		"keyword":    "match",
	}
)

//cate,shi-tang,shi-song|keyword,testword,word2
func CheckFilter(str string) (map[string][]interface{}, error) {
	maps := make(map[string][]interface{})
	if str == "" {
		return maps, nil
	}
	s := strings.Split(str, "|")
	for _, v := range s {
		columns := strings.Split(v, ",")
		field := columns[0]
		if field == "" {
			return maps, errors.New(v + " Field is empty")
		}
		if _, ok := _filterMaps[field]; ok {
			tmp := make([]interface{}, 0)
			for i := 1; i < len(columns); i++ {
				if columns[i] == "" {
					continue
				}
				tmp = append(tmp, columns[i])
			}
			maps[field] = tmp
		} else {
			return maps, errors.New(v + " Field params is invalid")
		}
	}

	return maps, nil

}

func (d *Dao) EsSearch(c context.Context, req *model.Query) ([]*model.Poem, int64, error) {
	var (
		poems = make([]*model.Poem, 0)
		maps  = make(map[string][]interface{})
		err   error
	)
	if maps, err = CheckFilter(req.Filter); err != nil {
		return poems, 0, ecode.Error(ecode.RequestErr, err.Error())
	}

	query := elastic.NewBoolQuery()
	if val, ok := maps["keyword"]; ok {
		keyWordQuery := elastic.NewMultiMatchQuery(val[0]).
			FieldWithBoost("paragraphs", 4).
			FieldWithBoost("content", 4).
			FieldWithBoost("rhythmic", 4).
			FieldWithBoost("notes", 3).
			FieldWithBoost("chapter", 2).
			FieldWithBoost("section", 2).
			FieldWithBoost("comment", 2).
			FieldWithBoost("title", 2)
		query = query.Must(keyWordQuery)
	}
	delete(maps, "keyword")
	for k, v := range maps {
		if _filterMaps[k] == "match" {
			query = query.Must(elastic.NewMatchQuery(k, v[0]))
		} else {
			query = query.Filter(elastic.NewTermsQuery(k, v...))
		}
	}
	if log.GetLogConf().Level == zap.DebugLevel.String() {
		source, err := query.Source()
		log.SugarWithContext(c).Debugf("poems EsSearch Source:%#v\n Err:%#v\n", source, err)
	}

	//对筛选的诗词进行随机排序
	script := elastic.NewScript("Math.random()")
	scriptSort := elastic.NewScriptSort(script, "number").Asc()
	search := d.es.Search(model.POEMS_ES_INDEX).Query(query).
		From(int((req.PageNum - 1) * int64(req.PageSize))).SortBy(scriptSort).
		Size(int(req.PageSize))
	resp, err := search.Do(c)
	if err != nil {
		source, err := query.Source()
		log.SugarWithContext(c).Errorf("poems EsSearch  Err: Source:%#v\n Err:%#v\n", source, err)
		return poems, 0, errors.WithStack(err)
	}
	log.SugarWithContext(c).Debugf("poems EsSearch Res:%#v\n", resp)
	if resp != nil {
		log.SugarWithContext(c).Debugf("poems EsSearch Res.TotalHits:%#v\n", resp.TotalHits())
		log.SugarWithContext(c).Debugf("poems EsSearch Res.Hits.Hits:%#v\n", resp.Hits.Hits)
	}

	if resp.TotalHits() > 0 {
		for _, hit := range resp.Hits.Hits {
			var t model.Poem
			err := utils.JsonUnmarshal(hit.Source, &t)
			if err != nil {
				log.ZapWithContext(c).Error("ES Select Poems json.Unmarshal resp Err On EsSearch :" + err.Error() + string(hit.Source))
				continue
			}
			poems = append(poems, &t)
		}
	}

	return poems, resp.TotalHits(), errors.WithStack(err)
}
