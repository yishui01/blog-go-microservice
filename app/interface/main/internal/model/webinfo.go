package model

type FrontListWebInfo struct {
	PageComm
	Lists []*FrontWebInfoDetail
}

type FrontWebInfoDetail struct {
	Sn     string
	WebKey string
	WebVal string
}
