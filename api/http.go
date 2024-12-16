package api

import "net/http"

type HttpPkg struct {
	HttpPkgInterface
}

func (HttpPkg) Get(url string) (resp *http.Response, err error) {
	return http.Get(url)
}

var Http HttpPkg

type HttpPkgInterface interface {
	Get(url string) (resp *http.Response, err error)
}
