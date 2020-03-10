package es

import (
	"github.com/olivere/elastic/v7"
	"github.com/spf13/viper"
	"github.com/zuiqiangqishao/framework/pkg/utils"
)

type ESearchConf struct {
	Urls     []string
	UserName string
	Passwd   string
	Sniff    bool
}

func New(c *ESearchConf) *elastic.Client {
	if c == nil {
		c = setDefaultConf()
	}
	cf := []elastic.ClientOptionFunc{
		elastic.SetURL(c.Urls...),
		elastic.SetSniff(c.Sniff),
		elastic.SetHealthcheck(true),
	}
	if c.UserName != "" || c.Passwd != "" {
		cf = append(cf, elastic.SetBasicAuth(c.UserName, c.Passwd))
	}
	es, err := elastic.NewClient(cf...)
	if err != nil {
		utils.PanicErr(err)
	}
	return es
}

func setDefaultConf() *ESearchConf {
	c := new(ESearchConf)

	if err := viper.Sub("elastic").Unmarshal(c); err != nil {
		panic("unable to decode Elastic Config struct, %v" + err.Error())
	}

	return c
}
