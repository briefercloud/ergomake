package ginutils

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type bytesBinding struct{}

var BYTES = bytesBinding{}

func (bytesBinding) Name() string {
	return "bytes"
}

// Bind reads the body of the http.Request and copies it into obj
func (bytesBinding) Bind(req *http.Request, obj any) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read request body")
	}
	defer req.Body.Close()

	return bytesBinding{}.BindBody(body, obj)
}

// BindBody copies the given body into obj
func (bytesBinding) BindBody(body []byte, obj any) error {
	bytesObj, ok := obj.(*[]byte)
	if !ok {
		return errors.New("obj must be a pointer to []byte")
	}
	*bytesObj = make([]byte, len(body))
	copy(*bytesObj, body)
	return nil
}
