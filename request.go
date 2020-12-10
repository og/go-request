package greq

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RequestQuery interface {
	Query() (url.Values, error)
}
type RequestFormUrlencoded interface {
	FormUrlencoded() (url.Values, error)
}
type RequestFormData interface {
	FormData(bufferData *bytes.Buffer) (*multipart.Writer, error)
}
type RequestHeader interface {
	Header() (http.Header, error)
}

type Request struct {
	Query RequestQuery
	FormUrlencoded RequestFormUrlencoded
	FormData RequestFormData
	Header RequestHeader
	JSON []byte
}

func HttpMessage(r http.Request) string {
	data, err := httputil.DumpRequest(&r, true) ; if err != nil {panic(err)}
	data = bytes.TrimSuffix(data, []byte("\r\n\r\n"))
	return string(data)
}
