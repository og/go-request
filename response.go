package greq

import (
	"net/http"
)

type responseBytes struct {
	Bind bool
	Bytes *[]byte
}
func BindBytes(ptr *[]byte) responseBytes {
	return responseBytes{
		Bind:  true,
		Bytes: ptr,
	}
}
type responseJSON struct {
	Bind bool
	Value interface{}
}
func BindJSON(ptr interface{}) responseJSON {
	return responseJSON{
		Bind:  true,
		Value: ptr,
	}
}

type httpResponse struct {
	Bind bool
	Value *http.Response
}
func HttpResponse(resp *http.Response) Response {
	return Response{
		httpResponse: resp,
	}
}
type Response struct {
	httpResponse *http.Response
	Bytes responseBytes
	JSON responseJSON
}
