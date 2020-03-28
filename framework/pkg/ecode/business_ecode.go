package ecode

//所有业务错误码在此文件定义

var (
	JsonErr        = add(100000)
	UpdateCacheErr = add(100001)
	UniqueErr      = add(100002)
)
