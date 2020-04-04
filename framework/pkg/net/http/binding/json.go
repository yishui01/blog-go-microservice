package binding

import (
	jsoniter "github.com/json-iterator/go"
	"net/http"

	"github.com/pkg/errors"
)

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj interface{}) error {
	decoder := jsoniter.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return errors.WithStack(err)
	}
	return validate(obj)
}
