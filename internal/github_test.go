package internal

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

type testDecoder struct {
	HttpPkgInterface
}

func (testDecoder) Get(url string) (resp *http.Response, err error) {
	return &http.Response{
		Status:           "",
		StatusCode:       0,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             http.NoBody,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil
}

type testResponse struct {
	HttpPkgInterface
}

var jsonStr = `[{
    "tag_name": "v3.26.0-beta"
  }]`

func (testResponse) Get(url string) (resp *http.Response, err error) {

	responseBody := io.NopCloser(bytes.NewReader([]byte(jsonStr)))
	return &http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil
}

type testError struct {
	HttpPkgInterface
}

func (testError) Get(url string) (resp *http.Response, err error) {
	return &http.Response{
		StatusCode: 404,
	}, errors.New("not found")
}

func Test_Github(t *testing.T) {
	_, err := GetGoAlgorandRelease("beta", new(testDecoder))
	if err == nil {
		t.Error("should fail to decode")
	}

	r, err := GetGoAlgorandRelease("beta", new(testResponse))
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Error("should not be nil")
	}
	if *r != "v3.26.0-beta" {
		t.Error("should return v3.26.0-beta")
	}

	_, err = GetGoAlgorandRelease("beta", new(testError))
	if err == nil {
		t.Error("should fail to get")
	}
}
