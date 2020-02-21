package utils

func PanicErr(err error) {
	if err != nil {
		panic("PANIC ERR:" + err.Error())
	}
}
