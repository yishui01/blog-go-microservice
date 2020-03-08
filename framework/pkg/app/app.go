package app

import (
	"os"
)

var (
	AppConf *AppConfig = DefaultAppConfig()
)

type AppConfig struct {
	AppName  string
	HostName string
}

func DefaultAppConfig() *AppConfig {
	h, _ := os.Hostname()
	return &AppConfig{
		AppName:  "unknowApp",
		HostName: h,
	}
}
