package model

const TIME_LAYOUT = "2006-01-02 15:04:05"

//返回response中的page数据
type PageComm struct {
	Total    int64
	Page     int64
	PageSize int32
}
