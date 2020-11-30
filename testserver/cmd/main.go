package main

import (
	greq "github.com/og/go-request"
	"github.com/og/go-request/testserver"
	"net/http"
)

func main () {
	testserver.Add(testserver.Data{
		Method: "GET",
		Path:  "/TestGet",
		Func: func(w http.ResponseWriter, r *http.Request) {
			testserver.Send(w, greq.HttpMessage(*r))
		},
	})
	testserver.Run()
}
