package app

import (
	"os"
	"sync"
)

var (
	_defaultAppConf *AppConfig = _unknowAppConfig()
	_mx             sync.RWMutex
)

type AppConfig struct {
	AppName  string
	HostName string
}

func GetAppConf() *AppConfig {
	_mx.RLock()
	defer _mx.RUnlock()
	t := *_defaultAppConf
	return &t
}

func SetAppConf(c *AppConfig) {
	_mx.RLock()
	defer _mx.RUnlock()
	_defaultAppConf = c
}

func _unknowAppConfig() *AppConfig {
	h, _ := os.Hostname()
	return &AppConfig{
		AppName:  "unknowApp",
		HostName: h,
	}
}
