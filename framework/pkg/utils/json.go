package utils

import jsoniter "github.com/json-iterator/go"

func JsonUnmarshal(input []byte, data interface{}) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(input, data)
}

func JsonMarshal(data interface{}) ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(data)
}
