package utils

import "fmt"

func PanicErr(err error) {
	if err != nil {
		panic("PANIC ERR:" + err.Error())
	}
}

func Vf(s interface{}) string {
	return fmt.Sprintf("%#+v", s)
}
