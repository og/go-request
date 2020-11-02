package greq

import (
	"net/http"
	"net/http/httputil"
)

type Request struct {
	Query interface{}
	FormUrlencoded interface{}
	FormData interface{}
	// url.Values{}
	Header interface{}
	JSON interface{}
}

func HttpMessage(r http.Request) string {
	data, err := httputil.DumpRequestOut(&r, true) ; if err != nil {panic(err)}
	return string(data)
}
