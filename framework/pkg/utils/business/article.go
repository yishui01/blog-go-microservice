package business

import "regexp"

var (
	_ascReg   = regexp.MustCompile(`^(.+)\|(asc|desc)$`)
	_orderKey = map[string]bool{
		"created_at": true,
		"updated_at": true,
		"view_count": true,
		"cm_count":   true,
		"laud_count": true,
	}
)

func ArtOrderReg() *regexp.Regexp {
	return _ascReg
}

func ArtOrderKey() map[string]bool {
	return _orderKey
}
