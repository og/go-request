package greq

import (
	"bytes"
	"fmt"
	ge "github.com/og/x/error"
	core_ogjson "github.com/og/x/json/core"
	"io"
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

func (c Client) Get(URL string, request Request) (res Response) {
	return c.Send(GET, URL, request)
}
func (c Client) Post(URL string, request Request) (res Response) {
	return c.Send(POST, URL, request)
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
func (c Client) Send(method Method, URL string, request Request) (resp Response) {

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
	var formWriter *multipart.Writer
	if formData := request.FormData; formData != nil {
		rForm := reflect.ValueOf(formData)
		rFormType := rForm.Type()
		bufferData := bytes.NewBuffer(nil)
		formWriter = multipart.NewWriter(bufferData)
		for i:=0;i<rFormType.NumField();i++ {
			itemType := rFormType.Field(i)
			itemValue := rForm.Field(i)
			fieldName := ""
			if value, has := itemType.Tag.Lookup("form") ; has {
				fieldName = value
			} else {
				fieldName = itemType.Name
			}
			if itemValue.Type().String() == "*os.File" {
				file := itemValue.Interface().(*os.File)
				fileW, err := formWriter.CreateFormFile(fieldName, file.Name())
				_, err = io.Copy(fileW, file) ; ge.Check(err)
			} else {
				err := formWriter.WriteField(fieldName, fmt.Sprintf("%v", itemValue.Interface())) ; ge.Check(err)
			}
		}
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
	resp.HttpResponse = httpResp
	return
}