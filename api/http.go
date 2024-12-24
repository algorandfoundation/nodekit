package api

import (
	"io"
	"net/http"
)

type HttpPkg struct {
	HttpPkgInterface
}

func (HttpPkg) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}
func (HttpPkg) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return http.Post(url, contentType, body)
}

var Http HttpPkg

type HttpPkgInterface interface {
	Get(url string) (resp *http.Response, err error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}
