package greq_test

import (
	"bytes"
	"context"
	greq "github.com/og/go-request"
	"github.com/og/go-request/testserver"
	gjson "github.com/og/json"
	gconv "github.com/og/x/conv"
	ge "github.com/og/x/error"
	gtest "github.com/og/x/test"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init () {
	go testserver.Run()
}
func hosturl(path string) string {
	return "http://127.0.0.1:2421" + path
}
func formatMessage(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "\r\n")
	s = strings.TrimSpace(s)
	return s
}
func formatFormDataMessage(respString string) string {
	rBoundary := regexp.MustCompile(`boundary=.{60}`)
	boundary := strings.Replace(rBoundary.FindString(respString), "boundary=", "", 1)
	return strings.ReplaceAll(respString, boundary, "testboundarytestboundarytestboundarytestboundarytestboundary")
}

type GetQuery struct {
	ID string
}
func (q GetQuery) Query() (url.Values, error) {
	v := url.Values{}
	v.Set("id", q.ID)
	return v, nil
}
func TestGet(t *testing.T) {
	as := gtest.NewAS(t)
	c := greq.New(greq.Config{})
	data := testserver.Data{
		Method: "GET",
		Path:  "/TestGet",
		Func: func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, greq.HttpMessage(*r))
		},
	}
	testserver.Add(data)
	query := GetQuery{
		ID: "a",
	}
	{
		var respBytes []byte
		err := c.Send(context.TODO(), data.Method, hosturl(data.Path), greq.Request{
			Query: query,
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
		if err != nil { panic(err)}
		as.Equal(string(respBytes), formatMessage(`
GET /TestGet?id=a HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
User-Agent: Go-http-client/1.1`))
	}
	{
		var respBytes []byte
		err := c.Send(context.TODO(), greq.POST, hosturl(data.Path), greq.Request{
			Query: query,
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		}) ; ge.Check(err)
		as.Equal(string(respBytes), "method is error: should be GET. request method is POST")
	}
}

type PostJSON struct {
	Name string `json:"name"`
}
func TestPost(t *testing.T) {
	as := gtest.NewAS(t)
	c := greq.New(greq.Config{})
	{
		data := testserver.Data{
			Method: "POST",
			Path:   "/TestPost",
			Func: func(w http.ResponseWriter, r *http.Request) {
				testserver.Send(w, greq.HttpMessage(*r))
			},
		}
		testserver.Add(data)
		var respBytes []byte
		err := c.Send(context.TODO(), greq.POST, hosturl(data.Path), greq.Request{
			JSON: PostJSON{Name: "nimoc"},
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		}) ; ge.Check(err)
		message := formatMessage(`
POST /TestPost HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Content-Length: 16
Content-Type: application/json
User-Agent: Go-http-client/1.1

{"name":"nimoc"}`)
		as.Equal(string(respBytes), message)
	}
}

func TestGetCookieJar(t *testing.T) {
	data := testserver.Data{
		Method: "GET",
		Path:   "/TestGetCookieJar",
		Func: func(w http.ResponseWriter, r *http.Request) {
			count := 0
			{
				requestCookie, err := r.Cookie("count")
				if err != nil {
					if err == http.ErrNoCookie {
						count = 0
					} else {
						panic(err)
					}
				} else {
					value := requestCookie.Value
					count = ge.Int(strconv.Atoi(value))
				}
			}
			count++
			cookie := &http.Cookie{
				Name: "count",
				Value: strconv.Itoa(count),
				Expires: time.Now().AddDate(0,0,1),
			}
			http.SetCookie(w, cookie)
			testserver.Send(w, gconv.IntString(count) + "\r\n" + "" + greq.HttpMessage(*r))
		},
	}
	testserver.Add(data)
	as := gtest.NewAS(t)

	cookieJar, err := cookiejar.New(nil) ; ge.Check(err)
	c := greq.New(greq.Config{
		HttpClient: &http.Client{
			Jar: cookieJar,
		},
	})
	{
		var respBytes []byte
		c.Send(context.TODO(), greq.Method(data.Method), hosturl(data.Path), greq.Request{}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})

		as.Equal(string(respBytes),
			formatMessage(`1
GET /TestGetCookieJar HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
User-Agent: Go-http-client/1.1`),
		)
	}
	{
		var respBytes []byte
		c.Send(context.TODO(), greq.Method(data.Method),hosturl(data.Path), greq.Request{}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
		as.Equal(
			string(respBytes),
			formatMessage(`2
GET /TestGetCookieJar HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Cookie: count=1
User-Agent: Go-http-client/1.1`),
	)
	}
	{
		var respBytes []byte
		c.Send(context.TODO(), greq.Method(data.Method),hosturl(data.Path), greq.Request{}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
		as.Equal(string(respBytes),
			formatMessage(`3
GET /TestGetCookieJar HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Cookie: count=2
User-Agent: Go-http-client/1.1`),
	)
	}
}

type Header struct {
	APIKey string
	User string
	Age int
	Float float64
	Bool bool
	List []string
}
func (h Header) Header () (http.Header, error) {
	header := http.Header{}
	header.Set("apiKey", h.APIKey)
	header.Set("user", h.User)
	return header, nil
}
func TestHeader(t *testing.T) {
	as := gtest.NewAS(t)
	c := greq.New(greq.Config{})
	data := testserver.Data{
		Method: "GET",
		Path:   "/TestHeader",
		Func: func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, gjson.String(r.Header))
		},
	}
	testserver.Add(data)
	header := Header {
		APIKey: "password",
		User: "nimoc",
		Age: 1,
		Float: 1.2,
		Bool: true,
		List: []string{"a", "c"},
	}
	var respBytes []byte
	err := c.Send(context.TODO(), greq.Method(data.Method), hosturl(data.Path), greq.Request{
		Header: header,
	}, greq.Response{
		Bytes: greq.BindBytes(&respBytes),
	})
	ge.Check(err)
	as.Equal(string(respBytes), formatMessage(`{"Accept-Encoding":["gzip"],"Apikey":["password"],"User":["nimoc"],"User-Agent":["Go-http-client/1.1"]}`))
}
type WWWForm struct {
	Name string
	Age int
}
func (f WWWForm) FormUrlencoded() (url.Values, error) {
	v := url.Values{}
	v.Set("name", f.Name)
	v.Set("age", gconv.IntString(f.Age))
	return v, nil
}
func TestWWWFormUrlencoded(t *testing.T) {
	as := gtest.NewAS(t)
	c := greq.New(greq.Config{})
	data := testserver.Data{
		Method: "GET",
		Path:   "/TestWWWFormUrlencoded",
		Func: func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, greq.HttpMessage(*r))
		},
	}
	testserver.Add(data)
	wwwForm := WWWForm {
		Name: "nimo",
		Age: 27,
	}
	{
		var respBytes []byte
		err := c.Send(context.TODO(), greq.Method(data.Method), hosturl(data.Path), greq.Request{
			FormUrlencoded: wwwForm,
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
		as.NoError(err)
		as.Equal(string(respBytes), formatMessage(`
GET /TestWWWFormUrlencoded HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Content-Length: 16
Content-Type: application/x-www-form-urlencoded
User-Agent: Go-http-client/1.1

age=27&name=nimo`))
	}
}


func TestJSON(t *testing.T) {
	as := gtest.NewAS(t)
	data := testserver.Data{
		Method: "POST",
		Path:   "/TestJSON",
		Func: func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, `{"name":"nimoc","age":18}`)
		},
	}
	testserver.Add(data)
	c := greq.New(greq.Config{})
	type RespData struct {
		Name string `json:"name"`
		Age int `json:"age"`
	}
	var respData RespData
	c.Send(context.TODO(), greq.Method(data.Method), hosturl(data.Path), greq.Request{}, greq.Response{
		JSON:  greq.BindJSON(&respData),
	})
	as.Equal(respData, RespData{
		Name: "nimoc",
		Age: 18,
	})
}
type UploadPhoto struct {
	Photo string
}
type FormData struct {
	Name string
	File string
	UploadPhoto
}
func (data FormData) FormData(bufferData *bytes.Buffer) (*multipart.Writer, error) {
	write := multipart.NewWriter(bufferData)
	var err error
	err = write.WriteField("name", data.Name) ; if err != nil {return nil, err}
	{
		fileField := "file"
		file, err := os.OpenFile(data.File,os.O_RDONLY,  0666) ; if err != nil {return nil, err}
		defer file.Close()
		fileWrite, err := write.CreateFormFile(fileField, file.Name()) ; if err != nil {return nil, err}
		_, err = io.Copy(fileWrite, file) ; if err != nil {return nil, err}
	}
	{
		fileField := "photo"
		file, err := os.OpenFile(data.Photo,os.O_RDONLY,  0666) ; if err != nil {return nil, err}
		defer file.Close()
		fileWrite, err := write.CreateFormFile(fileField, file.Name()) ; if err != nil {return nil, err}
		_, err = io.Copy(fileWrite, file) ; if err != nil {return nil, err}
	}
	return write, nil
}
func TestFormData(t *testing.T) {
	as := gtest.NewAS(t)
	_=as
	c := greq.New(greq.Config{})
	data := testserver.Data{
		Method: "POST",
		Path:   "/TestFormData",
		Func:  func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, greq.HttpMessage(*r))
		},
	}
	testserver.Add(data)
	form := FormData {
		Name: "nimo",
		File: "mock/json-1.json",
		UploadPhoto: UploadPhoto{
			"mock/json-1.json",
		},
	}
	var respBytes []byte
	err := c.Send(context.TODO(), greq.Method(data.Method), hosturl(data.Path), greq.Request{
		FormData: form,
	}, greq.Response{
		Bytes: greq.BindBytes(&respBytes),
	}) ; ge.Check(err)
	as.Equal(formatFormDataMessage(string(respBytes)), formatMessage(`
POST /TestFormData HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Content-Length: 654
Content-Type: multipart/form-data; boundary=testboundarytestboundarytestboundarytestboundarytestboundary
User-Agent: Go-http-client/1.1

--testboundarytestboundarytestboundarytestboundarytestboundary
Content-Disposition: form-data; name="name"

nimo
--testboundarytestboundarytestboundarytestboundarytestboundary
Content-Disposition: form-data; name="file"; filename="mock/json-1.json"
Content-Type: application/octet-stream

{"name": "nimoc","github": "http://github.com/nimoc"}
--testboundarytestboundarytestboundarytestboundarytestboundary
Content-Disposition: form-data; name="photo"; filename="mock/json-1.json"
Content-Type: application/octet-stream

{"name": "nimoc","github": "http://github.com/nimoc"}
--testboundarytestboundarytestboundarytestboundarytestboundary--`) + "\r\n")
}


type Full struct {
	Name string
	
}
func TestFull(t *testing.T) {
	as := gtest.NewAS(t)
	_=as

}