package greq_test

import (
	greq "github.com/og/go-request"
	"github.com/og/go-request/testserver"
	gconv "github.com/og/x/conv"
	ge "github.com/og/x/error"
	gjson "github.com/og/x/json"
	gtest "github.com/og/x/test"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init () {
	go testserver.Run()
}
func url(path string) string {
	return "http://127.0.0.1:2421" + path
}
func ExampleClient_Get() {
	c := greq.New(greq.Config{})
	{
		query := struct {
			ID string `query:"id"`
		}{
			ID: "a",
		}
		c.Get("http://www.github.com/", greq.Request{
			Query: query,
		})
	}
}
func ExampleClient_Response_JSON() {
	c := greq.New(greq.Config{})
	data := struct {
		Name string `json:"name"`
		Github string `json:"github"`
	}{}
	resp := c.Get("https://raw.githubusercontent.com/og/go-request/master/mock/json-1.json", greq.Request{})
	resp.JSON(&data)
}
func ExampleClient_Post() {
	c := greq.New(greq.Config{})
	data := struct {
		Name string `json:"name"`
		Github string `json:"github"`
	}{}
	resp := c.Post("https://raw.githubusercontent.com/og/go-request/master/mock/json-1.json", greq.Request{})
	resp.JSON(&data)
}
func formatMessage(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "\r\n")
	s = strings.TrimSpace(s)
	return s
}
func TestGet(t *testing.T) {
	as := gtest.NewAS(t)
	c := greq.New(greq.Config{})
	data := testserver.Data{
		Method: "GET",
		Path:   "/TestGet",
		Func: func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, greq.HttpMessage(*r))
		},
	}
	testserver.Add(data)
	query := struct {
		ID string `query:"id"`
	}{
		ID: "a",
	}
	{
		resp := c.Get(url(data.Path), greq.Request{
			Query: query,
		})
		as.Equal(resp.String(), formatMessage(`
GET /TestGet?id=a HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
User-Agent: Go-http-client/1.1`))
	}
	{
		resp := c.Post(url(data.Path), greq.Request{
			Query: query,
		})
		as.Equal(resp.String(), "method is error: should be GET. request method is POST")
	}
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
		resp := c.Post(url(data.Path), greq.Request{
			JSON: struct {
				Name string `json:"name"`
			}{Name: "nimoc"},
		})
		message := formatMessage(`
POST /TestPost HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Content-Length: 16
Content-Type: application/json
User-Agent: Go-http-client/1.1

{"name":"nimoc"}`)
		as.Equal(resp.String(), message)
		// 测试保留 Response.Body 数据，供后续代码使用
		as.Equal(as.NoErrorSecond(ioutil.ReadAll(resp.HttpResponse.Body)),  []byte(message))
		as.Equal(as.NoErrorSecond(ioutil.ReadAll(resp.HttpResponse.Body)),  []byte{})
	}
}

func Example_CookieJar() {
	cookieJar, err := cookiejar.New(nil) ; ge.Check(err)
	c := greq.New(greq.Config{
		HttpClient: &http.Client{
			Jar: cookieJar,
		},
	})
	c.Send(greq.GET, "http://github.com", greq.Request{})
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
	as.Equal(
		c.Send(greq.Method(data.Method),url(data.Path), greq.Request{}).String(),
		formatMessage(`1
GET /TestGetCookieJar HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
User-Agent: Go-http-client/1.1`),
		)
	as.Equal(
		c.Send(greq.Method(data.Method),url(data.Path), greq.Request{}).String(),
		formatMessage(`2
GET /TestGetCookieJar HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Cookie: count=1
User-Agent: Go-http-client/1.1`),
	)
	as.Equal(
		c.Send(greq.Method(data.Method),url(data.Path), greq.Request{}).String(),
		formatMessage(`3
GET /TestGetCookieJar HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
Cookie: count=2
User-Agent: Go-http-client/1.1`),
	)
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
	header := struct {
		APIKey string `header:"apikey"`
		User string `header:"user"`
		Age int
		Float float64
		Bool bool
		List []string
	}{
		APIKey: "password",
		User: "nimoc",
		Age: 1,
		Float: 1.2,
		Bool: true,
		List: []string{"a", "c"},
	}
	as.Equal(c.Send(greq.Method(data.Method), url(data.Path), greq.Request{
		Header: header,
	}).String(), formatMessage(`{"Accept-Encoding":["gzip"],"Age":["1"],"Apikey":["password"],"Bool":["true"],"Float":["1.2"],"List":["a","c"],"User":["nimoc"],"User-Agent":["Go-http-client/1.1"]}`))
}
// func TestWWWFormUrlencoded(t *testing.T) {
// 	as := gtest.NewAS(t)
// 	c := greq.New(greq.Config{})
// 	data := testserver.Data{
// 		Method: "GET",
// 		Path:   "/TestWWWFormUrlencoded",
// 		Func: func(w http.ResponseWriter, r *http.Request) {
// 			testserver.Send(w, greq.HttpMessage(*r))
// 		},
// 	}
// 	testserver.Add(data)
// 	wwwForm := struct {
// 		 Name string `form:"name"`
// 		 Age int `form:"age"`
// 	}{
//
// 	}
// 	c.Send(data.Method, url(data.Path), greq.Request{})
// 	as.Equal()
// }

func TestFormData(t *testing.T) {
	as := gtest.NewAS(t)
	c := greq.New(greq.Config{})
	data := testserver.Data{
		Method: "POST",
		Path:   "/TestFormData",
		Func:  func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, greq.HttpMessage(*r))
		},
	}
	testserver.Add(data)
	form := struct {
		Name string `form:"name"`
		File *os.File
	}{
		Name: "nimo",
		File: ge.File(os.OpenFile("mock/json-1.json",os.O_RDONLY,  0666)),
	}
	resp := c.Send(greq.Method(data.Method), url(data.Path), greq.Request{
		FormData: form,
	})
	as.Equal(resp.String(), "")
}