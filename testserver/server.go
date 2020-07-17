package testserver

import (
	ge "github.com/og/x/error"
	"net/http"
)

func Send(writer http.ResponseWriter, s string) {
	_, err := writer.Write([]byte(s)) ; ge.Check(err)
}
type Data struct {
	Method string
	Path string
	Func func(w http.ResponseWriter, r *http.Request)
}
func Run () {
	err := http.ListenAndServe(":2421", nil)
	if err != nil {
		panic(err)
	}
}
func Add(data Data) {
	http.HandleFunc(data.Path, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != data.Method {
			Send(writer, "method is error: should be " + data.Method + ". request method is " + request.Method)
			return
		}
		data.Func(writer, request)
	})
}