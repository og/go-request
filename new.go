package greq

import (
	"net/http"
)

type Config struct {
	HttpClient *http.Client
}
func New(config Config) (c Client) {
	if config.HttpClient != nil {
		c.HttpClient = config.HttpClient
	} else {
		c.HttpClient = &http.Client{}
	}
	return c
}