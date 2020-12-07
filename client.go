package greq

import (
	"bytes"
	"context"
	"errors"
	gjson "github.com/og/json"
	core_ogjson "github.com/og/json/core"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
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
func (c Client) Send(ctx context.Context, method Method, URL string, request Request, resp Response) error {
	var bodyReader io.Reader
	if request.JSON != nil {
		b, err := core_ogjson.Marshal(request.JSON, "json")
		if err != nil {return err}
		bodyReader = bytes.NewReader(b)
	}
	// x-www-form-urlencoded
	if request.FormUrlencoded != nil {
		values, err := request.FormUrlencoded.FormUrlencoded() ; if err != nil {return err}
		bodyReader = strings.NewReader(values.Encode())
	}
	// form data
	var formWriter *multipart.Writer
	if formData := request.FormData; formData != nil {
		bufferData := bytes.NewBuffer(nil)
		var err error
		formWriter, err = request.FormData.FormData(bufferData) ; if err != nil {return err}
		err = formWriter.Close() ; if err != nil {return err}
		bodyReader = bufferData
	}
	httpReq, err := http.NewRequestWithContext(ctx, string(method), URL, bodyReader) ; if err != nil {return err}
	// x-www-form-urlencoded
	if wwwFormUrlencoded := request.FormUrlencoded; wwwFormUrlencoded != nil {
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if request.FormData != nil {
		httpReq.Header.Add("Content-Type", formWriter.FormDataContentType())
	}
	// header
	if  request.Header != nil {
		header, err := request.Header.Header() ; if err != nil {return err}
		httpReq.Header = header
	}
	// query
	if request.Query != nil {
		values, err := request.Query.Query() ; if err != nil {return err}
		httpReq.URL.RawQuery = values.Encode()
	}
	// json
	if request.JSON != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpResp , err := c.HttpClient.Do(httpReq) ; ; if err != nil {return err}
	respBytes, err := ioutil.ReadAll(httpResp.Body) ; if err != nil {return err}
	err = httpResp.Body.Close() ; if err != nil {return err}
	if resp.Bytes.Bind {
		*resp.Bytes.Bytes = respBytes
	}
	if resp.JSON.Bind {
		if err := gjson.ParseBytesWithErr(respBytes, resp.JSON.Value); err != nil  {
			panic(errors.New(err.Error() + " source: " + string(respBytes)))
		}
	}
	return nil
}