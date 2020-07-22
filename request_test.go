package greq_test

import (
	greq "github.com/og/go-request"
	"github.com/og/go-request/testserver"
	gconv "github.com/og/x/conv"
	ge "github.com/og/x/error"
	gjson "github.com/og/x/json"
	gtest "github.com/og/x/test"
	"net/http"
	"net/http/cookiejar"
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
func url(path string) string {
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
		}, greq.Response{

		})
	}
}
func ExampleClient_Response_JSON() {
	c := greq.New(greq.Config{})
	data := struct {
		Name string `json:"name"`
		Github string `json:"github"`
	}{}
	c.Get("https://raw.githubusercontent.com/og/go-request/master/mock/json-1.json", greq.Request{}, greq.Response{
		JSON: greq.BindJSON(&data),
	})
}
func ExampleClient_Post() {
	c := greq.New(greq.Config{})
	data := struct {
		Name string `json:"name"`
		Github string `json:"github"`
	}{}
	c.Post("https://raw.githubusercontent.com/og/go-request/master/mock/json-1.json", greq.Request{}, greq.Response{
		JSON:  greq.BindJSON(&data),
	})
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
		var respBytes []byte
		c.Get(url(data.Path), greq.Request{
			Query: query,
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
		as.Equal(string(respBytes), formatMessage(`
GET /TestGet?id=a HTTP/1.1
Host: 127.0.0.1:2421
Accept-Encoding: gzip
User-Agent: Go-http-client/1.1`))
	}
	{
		var respBytes []byte
		c.Post(url(data.Path), greq.Request{
			Query: query,
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
		as.Equal(string(respBytes), "method is error: should be GET. request method is POST")
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
		var respBytes []byte
		c.Post(url(data.Path), greq.Request{
			JSON: struct {
				Name string `json:"name"`
			}{Name: "nimoc"},
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
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

func Example_CookieJar() {
	cookieJar, err := cookiejar.New(nil) ; ge.Check(err)
	c := greq.New(greq.Config{
		HttpClient: &http.Client{
			Jar: cookieJar,
		},
	})
	c.Send(greq.GET, "http://github.com", greq.Request{}, greq.Response{})
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
		c.Send(greq.Method(data.Method),url(data.Path), greq.Request{}, greq.Response{
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
		c.Send(greq.Method(data.Method),url(data.Path), greq.Request{}, greq.Response{
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
		c.Send(greq.Method(data.Method),url(data.Path), greq.Request{}, greq.Response{
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
	var respBytes []byte
	c.Send(greq.Method(data.Method), url(data.Path), greq.Request{
		Header: header,
	}, greq.Response{
		Bytes: greq.BindBytes(&respBytes),
	})
	as.Equal(string(respBytes), formatMessage(`{"Accept-Encoding":["gzip"],"Age":["1"],"Apikey":["password"],"Bool":["true"],"Float":["1.2"],"List":["a","c"],"User":["nimoc"],"User-Agent":["Go-http-client/1.1"]}`))
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
	wwwForm := struct {
		 Name string `form:"name"`
		 Age int `form:"age"`
	}{
		Name: "nimo",
		Age: 27,
	}
	{
		var respBytes []byte
		c.Send(greq.Method(data.Method), url(data.Path), greq.Request{
			FormUrlencoded: wwwForm,
		}, greq.Response{
			Bytes: greq.BindBytes(&respBytes),
		})
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

type UploadPhoto struct {
	File *os.File `form:"photo"`
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
	c.Send(greq.Method(data.Method), url(data.Path), greq.Request{}, greq.Response{
		JSON:  greq.BindJSON(&respData),
	})
	as.Equal(respData, RespData{
		Name: "nimoc",
		Age: 18,
	})
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
	form := struct {
		Name string `form:"name"`
		File *os.File `form:"file"`
		UploadPhoto
	}{
		Name: "nimo",
		File: ge.File(os.OpenFile("mock/json-1.json",os.O_RDONLY,  0666)),
		UploadPhoto: UploadPhoto{
			ge.File(os.OpenFile("mock/json-1.json",os.O_RDONLY,  0666)),
		},
	}
	var respBytes []byte
	c.Send(greq.Method(data.Method), url(data.Path), greq.Request{
		FormData: form,
	}, greq.Response{
		Bytes: greq.BindBytes(&respBytes),
	})
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
--testboundarytestboundarytestboundarytestboundarytestboundary--
`))

}