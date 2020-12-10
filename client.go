package greq

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
)

type Client struct {
	HttpClient *http.Client
}

type Method string
func (m Method) String() string {
	return string(m)
}
const GET Method = "GET"
const POST Method = "POST"
const HEAD Method = "HEAD"
const PUT Method = "PUT"
const DELETE Method = "DELETE"
const CONNECT Method = "CONNECT"
const OPTIONS Method = "OPTIONS"
const TRACE Method = "TRACE"
const PATCH Method = "PATCH"
func (c Client) Do(ctx context.Context, method Method, URL string, request Request) (resp *http.Response,  statusCode int, requestErr error) {
	// 防止对 resp nil 执行 resp.Body.Close()
	resp = httptest.NewRecorder().Result()
	// respClose = func() error {
	// 	return resp.Body.Close()
	// }
	var bodyReader io.Reader
	if request.JSON != nil {
		bodyReader = request.JSON
	}
	// x-www-form-urlencoded
	if request.FormUrlencoded != nil {
		values, err := request.FormUrlencoded.FormUrlencoded() ; if err != nil {return resp, 0, err}
		bodyReader = strings.NewReader(values.Encode())
	}
	// form data
	var formWriter *multipart.Writer
	if formData := request.FormData; formData != nil {
		bufferData := bytes.NewBuffer(nil)
		var err error
		formWriter, err = request.FormData.FormData(bufferData) ; if err != nil {return resp, 0, err}
		err = formWriter.Close() ; if err != nil {return resp, 0, err}
		bodyReader = bufferData
	}
	httpReq, err := http.NewRequestWithContext(ctx, string(method), URL, bodyReader) ; if err != nil {return resp, 0, err}
	// x-www-form-urlencoded
	if wwwFormUrlencoded := request.FormUrlencoded; wwwFormUrlencoded != nil {
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if request.FormData != nil {
		httpReq.Header.Add("Content-Type", formWriter.FormDataContentType())
	}
	// header
	if  request.Header != nil {
		header, err := request.Header.Header() ; if err != nil {return resp, 0, err}
		httpReq.Header = header
	}
	// query
	if request.Query != nil {
		values, err := request.Query.Query() ; if err != nil {return resp, 0, err}
		httpReq.URL.RawQuery = values.Encode()
	}
	// json
	if request.JSON != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	resp , err = c.HttpClient.Do(httpReq) ; ; if err != nil {return resp, 0, err}
	return resp, resp.StatusCode, nil
}