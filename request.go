package greq

import (
	ge "github.com/og/x/error"
	glist "github.com/og/x/list"
	gmap "github.com/og/x/map"
	"io/ioutil"
	"net/http"
	"strings"
)

type Request struct {
	Query interface{}
	WWWFormUrlencoded interface{}
	// url.Values{}
	Header interface{}
	JSON interface{}
}

func HttpMessage(r http.Request) string {
	sList := glist.StringList{}
	sList.Push(r.Method, " ", r.RequestURI, " ", r.Proto, "\r\n")
	sList.Push("Host: ", r.Host, "\r\n")
	keys := gmap.UnsafeKeys(r.Header).String() ; _ = r.Header[keys[0]]
	for _, key := range keys {
		values := r.Header[key]
		sList.Push(key, ": ", strings.Join(values, ","), "\r\n")
	}
	sList.Push("\r\n")
	bodyBytes , err := ioutil.ReadAll(r.Body) ; ge.Check(err)
	sList.Push(string(bodyBytes))
	return strings.TrimSpace(sList.Join(""))
}