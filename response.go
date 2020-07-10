package greq

import (
	"bytes"
	"github.com/og/x/error"
	"github.com/og/x/json"
	"io/ioutil"
	"net/http"
)

type Response struct {
	HttpResponse *http.Response
}
func (resp Response) Bytes() []byte {
	r := resp.HttpResponse
	b, err := ioutil.ReadAll(r.Body) ; ge.Check(err)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(b)) // 保留 Response.Body 数据，供后续代码使用
	return b
}
func (resp Response) String() string {
	return string(resp.Bytes())
}
func (resp Response) JSON(v interface{}) {
	gjson.ParseBytes(resp.Bytes(), v)
}