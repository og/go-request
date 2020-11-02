package greq

import (
	"bytes"
	"fmt"
	ge "github.com/og/x/error"
	gjson "github.com/og/json"
	core_ogjson "github.com/og/json/core"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

type Client struct {
	HttpClient *http.Client
}

func (c Client) Get(URL string, request Request, response Response) {
	c.Send(GET, URL, request, response)
}
func (c Client) Post(URL string, request Request, response Response) {
	c.Send(POST, URL, request, response)
}
type Method string
const GET Method = "GET"
const POST Method = "POST"
const HEAD Method = "HEAD"
const PUT Method = "PUT"
const DELETE Method = "DELETE"
const CONNECT Method = "CONNECT"
const OPTIONS Method = "OPTIONS"
const TRACE Method = "TRACE"
const PATCH Method = "PATCH"
func (c Client) Send(method Method, URL string, request Request, resp Response) {

	var bodyReader io.Reader
	if request.JSON != nil {
		b, err := core_ogjson.Marshal(request.JSON, "json") ; ge.Check(err)
		bodyReader = bytes.NewReader(b)
	}
	// x-www-form-urlencoded
	if wwwFormUrlencoded := request.FormUrlencoded; wwwFormUrlencoded != nil {
		urlValues := url.Values{}
		for key, values := range structToMap(request.FormUrlencoded, "form") {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
		bodyReader = strings.NewReader(urlValues.Encode())
	}
	// form data
	var formWriter *multipart.Writer
	if formData := request.FormData; formData != nil {

		bufferData := bytes.NewBuffer(nil)
		formWriter = multipart.NewWriter(bufferData)
		scan(formData, func(value reflect.Value, field reflect.StructField) {
			fieldName := ""
			if value, has := field.Tag.Lookup("form") ; has {
				fieldName = value
			} else {
				fieldName = field.Name
			}
			if field.Type.String() == "*os.File" {
				file := value.Interface().(*os.File)
				fileW, err := formWriter.CreateFormFile(fieldName, file.Name())
				_, err = io.Copy(fileW, file) ; ge.Check(err)
				return
			}
			if field.Type.Kind() == reflect.Struct {
				return
			}
			if field.Type.Kind() == reflect.Slice {
				return
			}
			err := formWriter.WriteField(fieldName, fmt.Sprintf("%v", value.Interface())) ; ge.Check(err)
		})
		ge.Check(formWriter.Close())
		bodyReader = bufferData
	}
	httpReq, err := http.NewRequest(string(method), URL, bodyReader) ; ge.Check(err)
	// x-www-form-urlencoded
	if wwwFormUrlencoded := request.FormUrlencoded; wwwFormUrlencoded != nil {
		httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if request.FormData != nil {
		httpReq.Header.Add("Content-Type", formWriter.FormDataContentType())
	}
	// header
	if header := request.Header; header != nil {
		for key, values := range structToMap(header, "header") {
			for _, value := range values {
				httpReq.Header.Add(key, value)
			}
		}
	}
	// query
	if request.Query != nil {
		query := httpReq.URL.Query()
		switch request.Query.(type) {
		case url.Values:
			values := request.Query.(url.Values)
			for key, valueList := range values {
				for _, value := range  valueList {
					query.Add(key, value)
				}
			}
		default:
			data, err := core_ogjson.Marshal(request.Query, "query") ; ge.Check(err)
			queryMap := map[string]string{}
			err = core_ogjson.Unmarshal(data, &queryMap, "query") ; ge.Check(err)
			for key, value := range queryMap {
				query.Add(key, value)
			}
		}
		httpReq.URL.RawQuery = query.Encode()
	}
	// json
	if request.JSON != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpResp , err := c.HttpClient.Do(httpReq) ; ge.Check(err)
	defer func() {
		ge.Check(httpResp.Body.Close())
	}()
	respBytes, err := ioutil.ReadAll(httpResp.Body) ; ge.Check(err)
	if resp.Bytes.Bind {
		*resp.Bytes.Bytes = respBytes
	}
	if resp.JSON.Bind {
		gjson.ParseBytes(respBytes, resp.JSON.Value)
	}
	return
}