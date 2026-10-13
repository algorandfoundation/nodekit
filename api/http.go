package api

import (
	"io"
	"net/http"
)

type HttpPkg struct {
	HttpPkgInterface
}

// Get performs a GET carrying the nodekit User-Agent. It is http.Get with that
// one addition: http.Get builds its request internally, leaving no place to
// set a header.
func (HttpPkg) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	SetUserAgent(req)
	return http.DefaultClient.Do(req)
}

// Post performs a POST carrying the nodekit User-Agent, otherwise matching
// http.Post.
func (HttpPkg) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	SetUserAgent(req)
	return http.DefaultClient.Do(req)
}

var Http HttpPkg

type HttpPkgInterface interface {
	Get(url string) (resp *http.Response, err error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}
