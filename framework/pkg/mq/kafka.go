package mq

import "sync"

var (
	_defaultKafConf *KafConf = &KafConf{}
	_kmx            sync.RWMutex
)

type KafConf struct {
	Brokers []string
}

func GetKafkaConf() *KafConf {
	_kmx.RLock()
	defer _kmx.RUnlock()
	t := *_defaultKafConf
	return &t
}

func SetKafkaConf(c *KafConf) {
	_kmx.Lock()
	defer _kmx.Unlock()
	_defaultKafConf = c
}
